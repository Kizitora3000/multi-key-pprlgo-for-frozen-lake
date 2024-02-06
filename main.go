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

const EPISODES = 1000 // 1000000

const MAX_USERS = 2

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

	// PPRL
	for episode := 0; episode < EPISODES; episode++ {
		state := env.Reset()
		for {
			action := agt.ChooseRandomAction()
			next_state, reward, done := env.Step(action)
			agt.Learn(state, action, reward, next_state, testContext, encryptedQtable, user_list)

			if done {
				break
			}
			state = next_state

		}
	}

	agt.ShowQTable()
	agt.ShowOptimalPath(env)
	fmt.Println(calcMSE(agt, encryptedQtable, testContext))
	ShowDecryptedQTable(agt, encryptedQtable, testContext)
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
