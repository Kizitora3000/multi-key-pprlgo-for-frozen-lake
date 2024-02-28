package pprl

import (
	"MKpprlgoFrozenLake/mkckks"
	"MKpprlgoFrozenLake/utils"
)

func initializeZeros(Na int, params mkckks.Parameters) *mkckks.Message {
	zeros := mkckks.NewMessage(params)
	for i := 0; i < Na; i++ {
		zeros.Value[i] = complex(0, 0)
	}
	return zeros
}

func initializeOnes(Na int, params mkckks.Parameters) *mkckks.Message {
	ones := mkckks.NewMessage(params)
	for i := 0; i < Na; i++ {
		ones.Value[i] = complex(1, 0)
	}
	return ones
}

func SecureQtableUpdating(v_t []float64, w_t []float64, Q_new float64, Nv int, Na int, testContext *utils.TestParams, EncryptedQtable []*mkckks.Ciphertext, user_name string) {
	v_t_expanded := make([]*mkckks.Ciphertext, Nv)

	// 行動(w_t)は横の行列のため、縦の行列である状態(v_t)も横に拡張する
	//[0,		[0, 0, 0, 0]
	// 1,  ->   [1, 1, 1, 1]
	// 0]		[0, 0, 0, 0]
	for i := 0; i < Nv; i++ {
		if v_t[i] == 0 {
			zeros := initializeZeros(Na, testContext.Params)
			v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(zeros, testContext.PkSet.GetPublicKey(user_name))
		} else if v_t[i] == 1 {
			ones := initializeOnes(Na, testContext.Params)
			v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(ones, testContext.PkSet.GetPublicKey(user_name))
		}
	}

	w_t_msg := mkckks.NewMessage(testContext.Params)
	for i := 0; i < Na; i++ {
		w_t_msg.Value[i] = complex(w_t[i], 0) // 虚部は0
	}
	fhe_w_t := testContext.Encryptor.EncryptMsgNew(w_t_msg, testContext.PkSet.GetPublicKey(user_name))

	Q_news_msg := mkckks.NewMessage(testContext.Params)
	for i := 0; i < Na; i++ {
		Q_news_msg.Value[i] = complex(Q_new, 0) // 虚部は0
	}
	fhe_Q_news := testContext.Encryptor.EncryptMsgNew(Q_news_msg, testContext.PkSet.GetPublicKey(user_name))

	for i := 0; i < Nv; i++ {
		// EncryptedQtable[i] = EncryptedQtable[i] + Qnew * v_t * w_t - Qold * v_t * w_t

		fhe_v_t := v_t_expanded[i]

		// calc: Qnew * (v_t * w_t)
		fhe_v_and_w_Qnew := testContext.Evaluator.MulRelinNew(fhe_v_t, fhe_w_t, testContext.RlkSet)
		fhe_v_and_w_Qnew = testContext.Evaluator.MulRelinNew(fhe_v_and_w_Qnew, fhe_Q_news, testContext.RlkSet)

		// calc: Qold * (v_t * w_t)
		fhe_v_and_w_Qold := testContext.Evaluator.MulRelinNew(fhe_v_t, fhe_w_t, testContext.RlkSet)
		fhe_v_and_w_Qold = testContext.Evaluator.MulRelinNew(fhe_v_and_w_Qold, EncryptedQtable[i], testContext.RlkSet)

		// ノイズ増加を防ぐため復号して除去する
		decrypt_fhe_v_and_w_Qnew := testContext.Decryptor.Decrypt(fhe_v_and_w_Qnew, testContext.SkSet)
		re_fhe_v_and_w_Qnew := testContext.Encryptor.EncryptMsgNew(decrypt_fhe_v_and_w_Qnew, testContext.PkSet.GetPublicKey(user_name))
		decrypt_fhe_v_and_w_Qold := testContext.Decryptor.Decrypt(fhe_v_and_w_Qold, testContext.SkSet)
		re_fhe_v_and_w_Qold := testContext.Encryptor.EncryptMsgNew(decrypt_fhe_v_and_w_Qold, testContext.PkSet.GetPublicKey(user_name))

		// calc: EncryptedQtable[i] += Qnew * v_t * w_t
		EncryptedQtable[i] = testContext.Evaluator.AddNew(EncryptedQtable[i], re_fhe_v_and_w_Qnew)
		// calc: EncryptedQtable[i] -= Qold * v_t * w_t (EncryptedQtable[i] = EncryptedQtable[i] + Qnew * v_t * w_t - Qold * v_t * w_t)
		EncryptedQtable[i] = testContext.Evaluator.SubNew(EncryptedQtable[i], re_fhe_v_and_w_Qold)
	}
}

func SecureActionSelection(v_t []float64, Nv int, Na int, testContext *utils.TestParams, EncryptedQtable []*mkckks.Ciphertext, user_name string) *mkckks.Ciphertext {
	v_t_expanded := make([]*mkckks.Ciphertext, Nv)

	// 行動(w_t)は横の行列のため、縦の行列である状態(v_t)も横に拡張する
	//[0,		[0, 0, 0, 0]
	// 1,  ->   [1, 1, 1, 1]
	// 0]		[0, 0, 0, 0]
	for i := 0; i < Nv; i++ {
		if v_t[i] == 0 {
			zeros := initializeZeros(Na, testContext.Params)
			v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(zeros, testContext.PkSet.GetPublicKey(user_name))
		} else if v_t[i] == 1 {
			ones := initializeOnes(Na, testContext.Params)
			v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(ones, testContext.PkSet.GetPublicKey(user_name))
		}
	}

	actions_msg := mkckks.NewMessage(testContext.Params)
	actions := testContext.Encryptor.EncryptMsgNew(actions_msg, testContext.PkSet.GetPublicKey(user_name))
	for i := 0; i < Nv; i++ {
		// s_t[i] == 1: [1, ..., 1] * [Q1, ..., Qn] = [Q1, ..., Qn](s_t)
		// s_t[i] == 0: [0, ..., 0] * [Q1, ..., Qn] = [0 , ..., 0]
		v_t_expanded[i] = testContext.Evaluator.MulRelinNew(v_t_expanded[i], EncryptedQtable[i], testContext.RlkSet)

		// [0 , ..., 0] + ... + [Q1, ..., Qn] + ... + [0 , ..., 0] = [Q1, ..., Qn](s_t)
		actions = testContext.Evaluator.AddNew(actions, v_t_expanded[i])
	}

	return actions
}
