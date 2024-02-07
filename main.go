package main

import (
	"MKpprlgoFrozenLake/agent"
	"MKpprlgoFrozenLake/environment"
	"MKpprlgoFrozenLake/frozenlake"
	"MKpprlgoFrozenLake/mkckks"
	"MKpprlgoFrozenLake/mkrlwe"
	"MKpprlgoFrozenLake/utils"
	"fmt"

	"github.com/ldsec/lattigo/v2/ckks"
)

const (
	// エピソード数の目安：3x3 → 32, 4x4 → 128, 5x5 → 32768, 6x6 → ? (TODO)
	EPISODES  = 32
	MAX_USERS = 2
)

func main() {
	// --- set up for RL ---
	lake := frozenlake.FrozenLake3x3
	env := environment.NewEnvironment(lake)
	agt := agent.NewAgent(env)

	// --- set up for multi key ---
	ckks_params, err := ckks.NewParametersFromLiteral(utils.FAST_BUT_NOT_128) // utils.FAST_BUT_NOT_128, utils.PPRL_PARAMS
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
	encryptedQtable := make([]*mkckks.Ciphertext, agt.GetStateNum())
	for i := 0; i < agt.GetStateNum(); i++ {
		plaintext := mkckks.NewMessage(testContext.Params)
		for i := 0; i < (1 << testContext.Params.LogSlots()); i++ {
			plaintext.Value[i] = complex(agt.InitValQ, 0) // 虚部は0
		}

		ciphertext := testContext.Encryptor.EncryptMsgNew(plaintext, testContext.PkSet.GetPublicKey(user_list[0])) // user_list[0] = "cloud"
		encryptedQtable[i] = ciphertext
	}

	// ---PPRL ---
	goal_count := 0
	for episode := 0; episode < EPISODES; episode++ {
		// 学習の進捗率と迷路の成功率の表示
		progress := float64(episode) / float64(EPISODES) * 100
		goal_rate := float64(goal_count) / float64(EPISODES) * 100.0
		fmt.Printf("\rTraining Progress: %.1f%% (%d/%d), Total Goal Rate: %.2f%%, ", progress, episode, EPISODES, goal_rate)

		state := env.Reset()
		for {
			action := agt.EpsilonGreedyAction(state)
			next_state, reward, done := env.Step(action)
			agt.Learn(state, action, reward, next_state, testContext, encryptedQtable, user_list)

			if done {
				if next_state == env.GoalPos {
					goal_count++
				}
				break
			}
			state = next_state

		}

		if episode%4 == 0 {
			evaluateGreedyActionAtEpisodes(env, agt)
		}
	}
	fmt.Println()

	// その他デバッグ情報の表示
	agt.ShowQTable()
	// agt.ShowOptimalPath(env)
	// fmt.Println(calcMSE(agt, encryptedQtable, testContext))
	// ShowDecryptedQTable(agt, encryptedQtable, testContext)
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
		fmt.Printf("State %d: %f\n", i, decryptedValue)
	}
}

func evaluateGreedyActionAtEpisodes(env *environment.Environment, agt *agent.Agent) {
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
	fmt.Printf("Greedy Action Goal Rate: %.2f%%\n", goalRate)
}
