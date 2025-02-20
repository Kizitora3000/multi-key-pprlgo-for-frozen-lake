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
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/ldsec/lattigo/v2/ckks"
)

const (
	EPISODES   = 200
	MAX_USERS  = 1
	MAX_TRIALS = 1
)

// 各ユーザからサーバへ送信されるQ値の更新情報を管理するためのチャネル
type QvalueUpdateData struct {
	V_t    []float64 // 状態のバイナリベクトル
	W_t    []float64 // 行動のバイナリベクトル
	Qvalue float64
}

func main() {
	// 乱数を固定 (*)
	rand.Seed(0)
	/*
		* プログラム全体で乱数を固定したいため，グローバルな乱数生成器である rand.Seed を使用している．
		ただし，意図せず他のファイルの乱数を固定してしまうことを避けるため，本来はローカルな乱数生成器である rand.New(rand.NewSource(0)) を使用することが推奨されている．
		今回はプログラム全体で乱数を固定したいので、rand.Seedを使用する．
	*/

	// -s フラグから氷結湖問題のサイズを取得
	map_size, is_measure := parseFlag()

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
	case "":
		log.Fatalf("error: the -s option is required")
	default:
		log.Fatalf("error: please choose from 3x3, 4x4, 5x5, 6x6")
	}

	// 処理時間計測用
	var elapsed_list []time.Duration

	// 成功率を算出するための変数を定義する．
	var success_rate_per_trial [][]float64
	var success_rate_per_trial_lock sync.Mutex
	var wg_trial sync.WaitGroup
	for trial := 0; trial < MAX_TRIALS; trial++ {
		wg_trial.Add(1)
		go func(trial int) {
			defer wg_trial.Done()

			// ---------- set up for multi key ----------

			ckks_params, err := ckks.NewParametersFromLiteral(utils.PN15QP880) // utils.FAST_BUT_NOT_128, utils.PN15QP880 (pprlと同じパラメータ)
			if err != nil {
				panic(err)
			}

			params := mkckks.NewParameters(ckks_params)
			user_list := make([]string, MAX_USERS+1) // MAX_USERS + "cloud platform"
			idset := mkrlwe.NewIDSet()

			user_list[0] = "cloud platform"

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

			// ---------- set up for RL ----------

			environments := make([]*environment.Environment, MAX_USERS)
			agents := make([]*agent.Agent, MAX_USERS)

			// init each environment and agent
			for user_i := 0; user_i < MAX_USERS; user_i++ {
				environments[user_i] = environment.NewEnvironment(lake)
				agents[user_i] = agent.NewAgent(environments[user_i])
				agents[user_i].QtableReset(environments[user_i])
				agents[user_i].Env.Reset()
			}

			// ---------- set up for PPRL ----------

			// クラウドプラットフォームの暗号化されたQテーブルを作成する．
			// 各エージェントの状態数・行動数は同一のためいずれのagentsを用いて初期化しても問題ないが，今回は代表としてagents[0]のQテーブルに基づいて作成する．
			encryptedQtable := encryptQtable(agents[0].Qtable, testContext, user_list[0]) // user_list[0] = "cloud platform"

			// 各ユーザからサーバへ送信されるQ値の更新情報を管理するためのチャネルを作成する．
			updateChannel := make(chan QvalueUpdateData, MAX_USERS)

			goal_count := 0
			total_espisode := 1
			var success_rate_per_episode = make([]float64, EPISODES+1) // episode = 1 からスタートする

			// 学習開始
			for total_espisode < EPISODES {
				var wg sync.WaitGroup

				for user_i := 0; user_i < MAX_USERS; user_i++ {
					// 各ユーザに独立したデータを渡すためにコピーを作成する．
					localTestContext := testContext.Copy()
					copiedEncryptedQtable := make([]*mkckks.Ciphertext, len(encryptedQtable))
					copy(copiedEncryptedQtable, encryptedQtable)

					wg.Add(1)
					go func(user_i int, copiedEncryptedQtable []*mkckks.Ciphertext, localTestContext *utils.TestParams) {
						defer wg.Done()

						env := environments[user_i]
						agt := agents[user_i]

						// 1ステップごとにユーザとクラウドプラットフォームのQテーブルを同期する．
						agt.Qtable = decryptQtable(encryptedQtable, localTestContext)

						state := agt.Env.AgentState

						action := agt.EpsilonGreedyAction(state)

						next_state, reward, done := env.Step(action)
						v_t, w_t, Q := agt.Trajectory(state, action, reward, next_state, copiedEncryptedQtable)

						updateChannel <- QvalueUpdateData{V_t: v_t, W_t: w_t, Qvalue: Q}

						state = next_state

						if done {
							if user_i == 0 {
								if state == env.GoalPos {
									goal_count++
								}

								total_espisode++
							}
							agt.Env.Reset()
						}
					}(user_i, copiedEncryptedQtable, localTestContext)
				}
				wg.Wait()

				goal_rate := float64(goal_count) / float64(total_espisode)
				success_rate_per_episode[total_espisode] = goal_rate

				// 各ユーザからの更新情報に基づいてクラウドプラットフォームのQテーブルを更新する．
				for user_i := 0; user_i < MAX_USERS; user_i++ {
					var start time.Time
					if is_measure {
						start = time.Now()
					}
					updateData := <-updateChannel
					pprl.SecureQtableUpdating(updateData.V_t, updateData.W_t, updateData.Qvalue, testContext, encryptedQtable, user_list[user_i+1])

					if is_measure {
						elapsed := time.Since(start)
						elapsed_list = append(elapsed_list, elapsed)
					}
				}

				// 平均処理時間を算出し、終了予測時刻を計算
				elapsed_sum := time.Duration(0)
				for i := 0; i < len(elapsed_list); i++ {
					elapsed_sum += elapsed_list[i]
				}
				elapsed_average := elapsed_sum / time.Duration(len(elapsed_list))

				// 現在の進捗状況から残りの処理時間を予測
				remaining_cnt := 5*MAX_USERS - len(elapsed_list)
				expected_total_time := elapsed_average * time.Duration(remaining_cnt)
				estimated_end_time := time.Now().Add(expected_total_time)

				// 残り時間を計算
				remaining_time := expected_total_time
				remaining_minutes := int(remaining_time.Minutes())
				remaining_seconds := int(remaining_time.Seconds()) % 60

				// trial > 2 以上の場合，代表として trial=0 の進捗を表示
				if trial == 0 && len(elapsed_list) <= 5*MAX_USERS {
					fmt.Printf("\r進捗:%5.1f%% (episode: %d/%d, max trial: %d), 平均処理時間: %s (%d個), 予測終了時刻: %s (残り %d分%d秒)",
						float64(total_espisode)/float64(EPISODES)*100,
						total_espisode,
						EPISODES,
						trial+1,
						elapsed_average,
						len(elapsed_list),
						estimated_end_time.Format("15:04:05"),
						remaining_minutes,
						remaining_seconds)
				} else {
					fmt.Println("end")
				}
			}

			// 各試行終了時にsuccess_rate_per_episodeのコピーを作成して追加
			success_rate_per_trial_lock.Lock()
			success_rate_per_trial = append(success_rate_per_trial, success_rate_per_episode)
			success_rate_per_trial_lock.Unlock()
		}(trial)
	}
	wg_trial.Wait()

	// 平均成功率をCSVに書き出す
	average_success_rate_filename := fmt.Sprintf("MKPPRL_average_success_rate_%dx%d_in_userNum_%d.csv", lake.Height, lake.Width, MAX_USERS)
	average_success_file, err := os.Create(average_success_rate_filename)
	if err != nil {
		panic(err)
	}
	defer average_success_file.Close()

	average_success_writer := csv.NewWriter(average_success_file)
	defer average_success_writer.Flush()

	// ヘッダーを書き込む
	average_success_writer.Write([]string{"Episode", "Average Success Rate"})

	// データを書き込む
	for episode := 1; episode <= EPISODES; episode++ {
		average_success_rate := 0.0

		for trial := 0; trial < MAX_TRIALS; trial++ {
			average_success_rate += success_rate_per_trial[trial][episode] / float64(MAX_TRIALS)
		}

		average_success_writer.Write([]string{fmt.Sprintf("%d", episode), fmt.Sprintf("%.2f", average_success_rate)})
	}
}

// -s フラグ (マップサイズの指定) を解析
func parseFlag() (string, bool) {
	// -s フラグを定義
	map_size := flag.String("s", "", "Size of the Frozen Lake map (options: 4x4, 5x5, 6x6)")
	is_measure := flag.Bool("m", false, "Set to true to measure execution time.")

	flag.Parse()

	return *map_size, *is_measure
}

func encryptQtable(qtable [][]float64, testContext *utils.TestParams, user_name string) []*mkckks.Ciphertext {
	N_state := len(qtable)
	N_action := len(qtable[0])

	encryptedQtable := make([]*mkckks.Ciphertext, N_state)

	for i := 0; i < N_state; i++ {
		plaintext := mkckks.NewMessage(testContext.Params)
		for i := 0; i < N_action; i++ {
			plaintext.Value[i] = complex(0, 0) // Q値は0で初期化
		}
		ciphertext := testContext.Encryptor.EncryptMsgNew(plaintext, testContext.PkSet.GetPublicKey(user_name))
		encryptedQtable[i] = ciphertext
	}

	return encryptedQtable
}

func decryptQtable(encryptedQtable []*mkckks.Ciphertext, testContext *utils.TestParams) [][]float64 {
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
