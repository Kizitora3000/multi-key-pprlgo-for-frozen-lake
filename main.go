package main

import (
	"MKpprlgoFrozenLake/agent"
	"MKpprlgoFrozenLake/environment"
	"MKpprlgoFrozenLake/frozenlake"
	"MKpprlgoFrozenLake/mkckks"
	"MKpprlgoFrozenLake/mkrlwe"
	"MKpprlgoFrozenLake/pprl"
	"MKpprlgoFrozenLake/utils"
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"

	"github.com/ldsec/lattigo/v2/ckks"
)

const (
	EPISODES   = 200
	MAX_USERS  = 3
	MAX_TRIALS = 100
)

func parseFlags() (string, error) {
	// コマンドライン引数でマップのサイズを指定
	mapSize := flag.String("s", "", "Size of the Frozen Lake map (options: 4x4, 5x5, 6x6)")
	flag.Parse()

	// `s`オプションが指定されているかチェック。指定されていなければエラーを返す
	if *mapSize == "" {
		return "", fmt.Errorf("Error: The -s option is required.")
	}

	return *mapSize, nil
}

func main() {
	// 乱数を固定
	// 補足：プログラム全体で乱数を固定したいためグローバルな乱数生成器である rand.Seed を使用している．のため、意図せず他のプログラムの乱数生成も固定してしまう可能性がある
	// 　　　ただし，意図せず他のファイルの乱数を固定してしまうことを避けるため，本来はローカルな乱数生成器である rand.New(rand.NewSource(0)) を使用することが推奨されている．
	// 　　　今回はプログラム全体で乱数を固定したいので、rand.Seedを使用する
	rand.Seed(0)

	// コマンドライン引数でマップのサイズを指定
	map_size, err := parseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var lake frozenlake.FrozenLake
	switch map_size {
	case "3x3":
		lake = frozenlake.FrozenLake3x3
	case "4x4":
		lake = frozenlake.FrozenLake4x4
	case "5x5":
		lake = frozenlake.FrozenLake5x5
	case "6x6":
		lake = frozenlake.FrozenLake6x6
	default:
		fmt.Println("Invalid map size. Please choose from 3x3, 4x4, 5x5, or 6x6.")
		os.Exit(1)
	}

	// --- set up for RL ---
	environments := make([]*environment.Environment, MAX_USERS)
	agents := make([]*agent.Agent, MAX_USERS)

	for i := 0; i < MAX_USERS; i++ {
		environments[i] = environment.NewEnvironment(lake)
		agents[i] = agent.NewAgent(environments[i])
	}

	// --- set up for multi key ---
	ckks_params, err := ckks.NewParametersFromLiteral(utils.FAST_BUT_NOT_128) // utils.FAST_BUT_NOT_128, utils.PN15QP880 (pprlと同じパラメータ)
	if err != nil {
		panic(err)
	}

	params := mkckks.NewParameters(ckks_params)
	user_list := make([]string, MAX_USERS+1) // AGENTS + "Cloud Platform"
	idset := mkrlwe.NewIDSet()

	user_list[0] = "cloud"
	// MAX_USERS分のIDを登録
	for i := 1; i <= MAX_USERS; i++ {
		user_list[i] = fmt.Sprintf("user%d", i)
	}

	for i := range user_list {
		idset.Add(user_list[i])
	}

	var testContext *utils.TestParams
	if testContext, err = utils.GenTestParams(params, idset); err != nil {
		panic(err)
	}

	// ---PPRL ---
	// 試行ごとにクラウドのQ値を初期化
	// 各エージェントの状態数・行動数は同一のため、いずれのagentsを用いて初期化しても問題ない。今回は代表としてagents[0]を使用する
	encryptedQtable := make([]*mkckks.Ciphertext, agents[0].GetStateNum())
	for i := 0; i < agents[0].GetStateNum(); i++ {
		plaintext := mkckks.NewMessage(testContext.Params)
		for i := 0; i < (1 << testContext.Params.LogSlots()); i++ {
			plaintext.Value[i] = complex(agents[0].InitValQ, 0) // 虚部は0
		}

		ciphertext := testContext.Encryptor.EncryptMsgNew(plaintext, testContext.PkSet.GetPublicKey(user_list[0])) // user_list[0] = "cloud"
		encryptedQtable[i] = ciphertext
	}

	for agent_idx := 0; agent_idx < MAX_USERS; agent_idx++ {
		agents[agent_idx].QtableReset(environments[agent_idx])
		agents[agent_idx].Env.Reset()
	}

	type UpdateData struct {
		V_t []float64
		W_t []float64
		Q   float64
	}

	// UpdateData型のチャネルを作成
	updateChannel := make(chan UpdateData, MAX_USERS)

	goal_count := 0
	total_espisode := 1
	var success_rate_per_episode = make([]float64, EPISODES+10) // 配列外参照を防ぐため余裕を持っておく

	for total_espisode < EPISODES {
		var wg sync.WaitGroup

		for agent_idx := 0; agent_idx < MAX_USERS; agent_idx++ {
			// 各ゴルーチンで独自のtestContextを生成する
			localTestContext := testContext.Copy()
			// クラウドプラットフォームのQテーブルをコピー
			copiedEncryptedQtable := make([]*mkckks.Ciphertext, len(encryptedQtable))
			for i, ct := range encryptedQtable {
				copiedEncryptedQtable[i] = ct.CopyNew()
			}

			wg.Add(1)
			go func(agent_idx int, copiedEncryptedQtable []*mkckks.Ciphertext, localTestContext *utils.TestParams) {
				defer wg.Done()

				env := environments[agent_idx]
				agt := agents[agent_idx]

				agt.Qtable = decryptedQTable(encryptedQtable, localTestContext)

				state := agt.Env.AgentState
				// action := agt.EpsilonGreedyAction(state)
				action := agt.SecureEpsilonGreedyAction(state, localTestContext, copiedEncryptedQtable, user_list[agent_idx+1])
				next_state, reward, done := env.Step(action)
				v_t, w_t, Q := agt.Trajectory(state, action, reward, next_state, copiedEncryptedQtable)

				updateChannel <- UpdateData{V_t: v_t, W_t: w_t, Q: Q}

				if done {
					if agent_idx == 0 {
						if next_state == env.GoalPos {
							goal_count++
						}

						total_espisode++
					}

					agt.Env.Reset()
				}
				state = next_state
			}(agent_idx, copiedEncryptedQtable, localTestContext)
		}
		wg.Wait()

		goal_rate := float64(goal_count) / float64(total_espisode)
		success_rate_per_episode[total_espisode] = goal_rate
		fmt.Println(total_espisode, goal_rate, goal_count)
		for agent_idx := 0; agent_idx < MAX_USERS; agent_idx++ {
			e := environments[agent_idx]
			a := agents[agent_idx]

			updateData := <-updateChannel
			pprl.SecureQtableUpdating(updateData.V_t, updateData.W_t, updateData.Q, a.Env.Height()*a.Env.Width(), len(e.ActionSpace), testContext, encryptedQtable, user_list[agent_idx+1])
		}
		decryptedQTable(encryptedQtable, testContext)
	}

	// 平均成功率をCSVに書き出す
	average_successl_rate_filename := fmt.Sprintf("MKPPRL_success_rate_%dx%d_in_usernum_%d.csv", environments[0].Height(), environments[0].Width(), MAX_USERS)
	average_file, err := os.Create(average_successl_rate_filename)
	if err != nil {
		panic(err)
	}
	defer average_file.Close()

	average_writer := csv.NewWriter(average_file)
	defer average_writer.Flush()

	// ヘッダーを書き込む
	average_writer.Write([]string{"Episode", "Success Rate"})

	// データを書き込む
	for episode, average_success_rate := range success_rate_per_episode {
		// episode は1で始まる
		if episode == 0 {
			continue
		}

		average_writer.Write([]string{fmt.Sprintf("%d", episode), fmt.Sprintf("%.2f", average_success_rate)})
	}
}

func decryptedQTable(encryptedQtable []*mkckks.Ciphertext, testContext *utils.TestParams) [][]float64 {
	qtable := make([][]float64, len(encryptedQtable))

	for i, encryptedValue := range encryptedQtable {
		// ここで復号化プロセスを実行
		decryptedValue := testContext.Decryptor.Decrypt(encryptedValue, testContext.SkSet)
		qtable[i] = make([]float64, decryptedValue.Slots())

		for j := 0; j < decryptedValue.Slots(); j++ {
			qtable[i][j] = real(decryptedValue.Value[j])
		}
	}
	return qtable
}

func ShowDecryptedQTable(encryptedQtable []*mkckks.Ciphertext, testContext *utils.TestParams) {
	// 暗号化されたQテーブルの各要素を復号化して表示
	fmt.Println("Decrypted Qtable:")
	for i, encryptedValue := range encryptedQtable {
		// ここで復号化プロセスを実行
		decryptedValue := testContext.Decryptor.Decrypt(encryptedValue, testContext.SkSet)
		// 復号化された値を表示
		// fmt.Printf("State %d: %f\n", i, decryptedValue)
		// 復号された値を表示
		height := int(math.Sqrt(float64(len(encryptedQtable))))
		x := i % height
		y := i / height
		fmt.Printf("State [Y: %d, X: %d]: ↑: %f, ↓: %f, ←: %f, →: %f\n", y, x, real(decryptedValue.Value[0]), real(decryptedValue.Value[1]), real(decryptedValue.Value[2]), real(decryptedValue.Value[3]))
	}
}
