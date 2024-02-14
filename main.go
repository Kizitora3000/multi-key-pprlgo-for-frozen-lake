package main

import (
	"MKpprlgoFrozenLake/agent"
	"MKpprlgoFrozenLake/environment"
	"MKpprlgoFrozenLake/frozenlake"
	"MKpprlgoFrozenLake/mkckks"
	"MKpprlgoFrozenLake/mkrlwe"
	"MKpprlgoFrozenLake/utils"
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/ldsec/lattigo/v2/ckks"
)

const (
	EPISODES   = 200
	MAX_USERS  = 2             // MAX_USERS = cloud + agents
	MAX_AGENTS = MAX_USERS - 1 // agents = MAX_USERS - cloud
)

func main() {
	// agent.ChooseRandomAction()やagent.EpsilonGreedyAction()で使用する乱数値を固定
	// 補足：rand.Seedはグローバルな乱数生成器のため、意図せず他のプログラムの乱数生成も固定してしまう可能性がある
	// 　　　rand.New(rand.NewSource(0))はローカルな乱数生成器のため、プログラムごとに乱数生成を固定でき、他のプログラムに影響を与えない
	// 　　　しかし、今回はプログラム全体で乱数を固定したいので、rand.Seedを使用する
	rand.Seed(0)

	// コマンドライン引数でマップのサイズを指定
	map_size := flag.String("s", "", "Size of the Frozen Lake map (options: 4x4, 5x5, 6x6)")
	flag.Parse()

	// `s`オプションが指定されているかチェック。指定されていなければ終了
	if *map_size == "" {
		fmt.Println("Error: The -s option is required.")
		os.Exit(1)
	}

	var lake frozenlake.FrozenLake
	switch *map_size {
	case "3x3":
		lake = frozenlake.FrozenLake3x3
	case "4x4":
		lake = frozenlake.FrozenLake4x4
	case "5x5":
		lake = frozenlake.FrozenLake5x5
	case "6x6":
		lake = frozenlake.FrozenLake6x6
	default:
		fmt.Println("Invalid map size. Please choose from 4x4, 5x5, or 6x6.")
		os.Exit(1)
	}

	// --- set up for RL ---
	environments := make([]*environment.Environment, MAX_AGENTS)
	agents := make([]*agent.Agent, MAX_AGENTS)

	for i := 0; i < MAX_AGENTS; i++ {
		environments[i] = environment.NewEnvironment(lake)
		agents[i] = agent.NewAgent(environments[i])
	}

	// --- set up for Result
	success_rate_filename := fmt.Sprintf("MKPPRL_success_rate_%dx%d.csv", lake.Height, lake.Width)
	file, err := os.Create(success_rate_filename)
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"Episode", "Success Rate"}) // 表頭を記入

	eval_rate_filename := fmt.Sprintf("MKPPRL_eval_greedy_success_rate_%dx%d.csv", lake.Height, lake.Width)
	// ファイルが存在する場合は削除 (eval_success_rateはeval関数が呼ばれるたびに追記していく形式なので、プログラム開始時は削除する)
	if _, err := os.Stat(eval_rate_filename); err == nil {
		if err := os.Remove(eval_rate_filename); err != nil {
			panic(err)
		}
	}

	// --- set up for multi key ---
	ckks_params, err := ckks.NewParametersFromLiteral(utils.PN15QP880) // utils.FAST_BUT_NOT_128, utils.PN15QP880 (pprlと同じパラメータ)
	if err != nil {
		panic(err)
	}

	params := mkckks.NewParameters(ckks_params)
	user_list := make([]string, MAX_USERS)
	idset := mkrlwe.NewIDSet()

	user_list[0] = "cloud"
	user_list[1] = "user1"

	for i := range user_list {
		idset.Add(user_list[i])
	}

	var testContext *utils.TestParams
	if testContext, err = utils.GenTestParams(params, idset); err != nil {
		panic(err)
	}

	// クラウドのQ値を初期化
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

	// ---PPRL ---
	goal_count := 0.0
	all_agt_eps := 0 // 各エージェントの試行回数の総計
	for episode := 0; episode <= EPISODES; episode++ {
		// 学習の進捗率を表示
		progress := float64(episode) / float64(EPISODES) * 100
		fmt.Printf("\rTraining Progress: %.1f%% (%d/%d)", progress, episode, EPISODES)

		for agent_idx := 0; agent_idx < MAX_AGENTS; agent_idx++ {
			env := environments[agent_idx]
			agt := agents[agent_idx]

			state := env.Reset()
			for {
				// action := agt.EpsilonGreedyAction(state)
				action := agt.SecureEpsilonGreedyAction(state, testContext, encryptedQtable, user_list)

				next_state, reward, done := env.Step(action)
				agt.Learn(state, action, reward, next_state, testContext, encryptedQtable, user_list)

				if done {
					if next_state == env.GoalPos {
						goal_count++
					}
					all_agt_eps++

					break
				}
				state = next_state
			}

			// 成功率を算出してcsvに出力
			goal_rate := goal_count / float64(all_agt_eps)
			writer.Write([]string{fmt.Sprintf("%d", int(episode)), fmt.Sprintf("%.2f", goal_rate)})

			if episode%4 == 0 {
				evaluateGreedyActionAtEpisodes(episode, env, agt)
			}
		}
	}
	fmt.Println()

	// その他デバッグ情報の表示
	agents[0].ShowQTable()
	// agents[0].ShowOptimalPath(environments[0])
	// ShowDecryptedQTable(agents[0], encryptedQtable, testContext)
	// fmt.Println(calcMSE(agt, encryptedQtable, testContext))
}

func calcMSE(agt *agent.Agent, encryptedQtable []*mkckks.Ciphertext, testContext *utils.TestParams) float64 {
	// 復号化されたQテーブルを格納するための変数
	decryptedQtable := make([][]float64, agt.GetStateNum())

	// 復号化されたQテーブルの初期化
	for i := range decryptedQtable {
		decryptedQtable[i] = make([]float64, agt.GetActionNum())
	}

	// encryptedQtableの復号化
	for i, encryptedValue := range encryptedQtable {
		decryptedMessage := testContext.Decryptor.Decrypt(encryptedValue, testContext.SkSet)
		for j := 0; j < agt.GetActionNum(); j++ {
			decryptedQtable[i][j] = real(decryptedMessage.Value[j])
		}
	}

	// MSEの計算
	var mse float64
	for i := range agt.Qtable {
		for j := range agt.Qtable[i] {
			diff := agt.Qtable[i][j] - decryptedQtable[i][j]
			mse += diff * diff
		}
	}
	mse /= float64(agt.GetStateNum() * agt.GetActionNum())

	return mse
}

func ShowDecryptedQTable(agt *agent.Agent, encryptedQtable []*mkckks.Ciphertext, testContext *utils.TestParams) {
	// 暗号化されたQテーブルの各要素を復号化して表示
	fmt.Println("Decrypted Qtable:")
	for i, encryptedValue := range encryptedQtable {
		// ここで復号化プロセスを実行
		decryptedValue := testContext.Decryptor.Decrypt(encryptedValue, testContext.SkSet)
		// 復号化された値を表示
		// fmt.Printf("State %d: %f\n", i, decryptedValue)
		// 復号された値を表示
		height := int(math.Sqrt(float64(agt.GetStateNum())))
		x := i % height
		y := i / height
		fmt.Printf("State [Y: %d, X: %d]: %f\n", y, x, decryptedValue.Value)
	}
}

func evaluateGreedyActionAtEpisodes(now_episode int, env *environment.Environment, agt *agent.Agent) {
	// ファイルを追記モードで開く（ファイルが存在しない場合は新しく作成）
	eval_rate_filename := fmt.Sprintf("MKPPRL_eval_greedy_success_rate_%dx%d.csv", env.Height(), env.Width())
	file, err := os.OpenFile(eval_rate_filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// ファイルが空（新規作成されたばかり）の場合、ヘッダーを書き込む
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	if fileInfo.Size() == 0 {
		writer.Write([]string{"Episode", "Success Rate"}) // 表頭を記入
	}

	goal_count := 0 // エピソードでのゴール到達回数をカウント
	trials := 100   // 評価のために各エピソードを何回実行するか

	for i := 0; i < trials; i++ {
		state := env.Reset()
		cnt := 0
		for {
			action := agt.GreedyAction(state) // 学習済みのQテーブルを使用して最適な行動を選択
			next_state, _, done := env.Step(action)

			if done {
				if next_state == env.GoalPos {
					goal_count++
				}
				break
			}

			if cnt > 100 {
				break
			}

			state = next_state
			cnt++
		}
	}

	goalRate := float64(goal_count) / float64(trials) * 100.0
	// fmt.Printf("Greedy Action Goal Rate: %.2f%%\n", goalRate)

	writer.Write([]string{fmt.Sprintf("%d", int(now_episode)), fmt.Sprintf("%.2f", goalRate)})
}
