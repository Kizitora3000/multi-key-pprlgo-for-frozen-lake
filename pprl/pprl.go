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

func initializeWa(Na int, Wa []float64, params mkckks.Parameters) *mkckks.Message {
	ones := mkckks.NewMessage(params)
	for i := 0; i < Na; i++ {
		ones.Value[i] = complex(Wa[i], 0)
	}
	return ones
}

func initializeQvalue(Na int, Q float64, params mkckks.Parameters) *mkckks.Message {
	ones := mkckks.NewMessage(params)
	for i := 0; i < Na; i++ {
		ones.Value[i] = complex(Q, 0)
	}
	return ones
}

func SecureQtableUpdating(v_t []float64, w_t []float64, Q_new float64, testContext *utils.TestParams, EncryptedQtable []*mkckks.Ciphertext, user_name string) {
	Nv := len(v_t)
	Na := len(w_t)

	Q_v_t_expanded := make([]*mkckks.Ciphertext, Nv)
	for i := 0; i < Nv; i++ {
		if v_t[i] == 0 {
			zeros := initializeZeros(Na, testContext.Params)
			Q_v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(zeros, testContext.PkSet.GetPublicKey(user_name))
		} else if v_t[i] == 1 {
			ones := initializeQvalue(Na, Q_new, testContext.Params)
			Q_v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(ones, testContext.PkSet.GetPublicKey(user_name))
		}
	}

	v_t_expanded := make([]*mkckks.Ciphertext, Nv)

	/*
		行動(w_t)は行ベクトルのため、列ベクトルである状態(v_t)を行方向に拡張する
		v_t -> v_t_expanted
		[0, -> [0, 0, 0, 0]
		 :      :  :  :  :
		 0, -> [0, 0, 0, 0]
		 1, -> [1, 1, 1, 1]
		 0, -> [0, 0, 0, 0]
		 :      :  :  :  :
		 0] -> [0, 0, 0, 0]
	*/

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

	// Q_news_msg := mkckks.NewMessage(testContext.Params)
	// for i := 0; i < Na; i++ {
	// 	Q_news_msg.Value[i] = complex(Q_new, 0) // 虚部は0
	// }
	// fhe_Q_news := testContext.Encryptor.EncryptMsgNew(Q_news_msg, testContext.PkSet.GetPublicKey(user_name))

	for i := 0; i < Nv; i++ {
		// EncryptedQtable[i] = EncryptedQtable[i] + Qnew * v_t * w_t - Qold * v_t * w_t

		fhe_v_t := v_t_expanded[i]

		// calc: Qnew * (v_t * w_t)
		fhe_v_and_w_Qnew := testContext.Evaluator.MulRelinNew(Q_v_t_expanded[i], fhe_w_t, testContext.RlkSet)
		// fhe_v_and_w_Qnew = testContext.Evaluator.MulRelinNew(fhe_v_and_w_Qnew, fhe_Q_news, testContext.RlkSet)

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

func SecureQtableMultiUpdating2(datas []utils.QvalueUpdateData, testContext *utils.TestParams, EncryptedQtable []*mkckks.Ciphertext, user_name string) {
	// fmt.Println(datas)

	Nv := len(datas[0].V_t)
	Na := len(datas[0].W_t)
	Nuser := len(datas)

	v := make([][]float64, Nv)
	for state := 0; state < Nv; state++ {
		v[state] = make([]float64, Na)

		for user := 0; user < Nuser; user++ {
			if datas[user].V_t[state] == 1 {
				// v[state] を1で初期化
				for i := 0; i < Na; i++ {
					v[state][i] = 1
				}
			}
		}
	}

	w := make([][]float64, Nv)
	for state := 0; state < Nv; state++ {
		w[state] = make([]float64, Na)

		for user := 0; user < Nuser; user++ {
			if datas[user].V_t[state] == 1 {
				// W_t[i] == 1　となる部分のみに1を入れる
				for i := 0; i < Na; i++ {
					if datas[user].W_t[i] == 1 {
						w[state][i] = 1
					}
				}
			}
		}
	}

	Qw := make([][]float64, Nv)
	for state := 0; state < Nv; state++ {
		Qw[state] = make([]float64, Na)

		for user := 0; user < Nuser; user++ {
			if datas[user].V_t[state] == 1 {
				// w[state] を [0, ..., 0, Q, 0, ..., 0] で初期化 (QとなるのはW_t[i]=1となる部分)
				for i := 0; i < Na; i++ {
					Qw[state][i] += datas[user].W_t[i] * datas[user].Qvalue
				}
			}

			if datas[0].V_t[state] == 1 && datas[1].V_t[state] == 1 {
				for i := 0; i < Na; i++ {
					Qw[state][i] = ((datas[0].W_t[i] * datas[0].Qvalue) + (datas[1].W_t[i] * datas[1].Qvalue)) / 2.0
				}
			}
		}
	}

	for state := 0; state < Nv; state++ {
		// EncryptedQtable[state] = EncryptedQtable[state] + Qnew * v_t * w_t - Qold * v_t * w_t

		// calc: v_t * (Qnew * w_t)
		v_and_Qnew_w := make([]float64, Na)
		for action := 0; action < Na; action++ {
			v_and_Qnew_w[action] = v[state][action] * Qw[state][action]
		}

		// get Qold
		Qold_msg := testContext.Decryptor.Decrypt(EncryptedQtable[state], testContext.SkSet)
		Qold := make([]float64, Na)
		for action := 0; action < Na; action++ {
			Qold[action] = real(Qold_msg.Value[action])
		}

		DecryptedQ := make([]float64, Na)
		copy(DecryptedQ, Qold)

		// calc: v_t * (Qold * w_t)
		v_and_Qold_w := make([]float64, Na)
		for action := 0; action < Na; action++ {
			v_and_Qold_w[action] = v[state][action] * (Qold[action] * w[state][action])
		}

		// calc: v_t * w_t
		v_w_t := make([]float64, Na)
		for action := 0; action < Na; action++ {
			for user := 0; user < Nuser; user++ {
				if datas[user].V_t[state]*datas[user].W_t[action] == 1 {
					v_w_t[action] = 1
				}
			}
		}

		// EncryptedQtable[i] = EncryptedQtable[i] + v_t * (Qnew * w_t) - v_t * (Qold * w_t)

		// calc1: EncryptedQtable[i] += v_t * (Qnew * w_t)
		for action := 0; action < Na; action++ {
			DecryptedQ[action] += v_and_Qnew_w[action]
		}
		// calc2: EncryptedQtable[i] -= v_t * (Qold * w_t)
		for action := 0; action < Na; action++ {
			DecryptedQ[action] -= v_and_Qold_w[action]
		}

		DecryptedQ_msg := mkckks.NewMessage(testContext.Params)
		for action := 0; action < Na; action++ {
			DecryptedQ_msg.Value[action] = complex(DecryptedQ[action], 0) // 虚部は0
		}
		DecryptedQ_ct := testContext.Encryptor.EncryptMsgNew(DecryptedQ_msg, testContext.PkSet.GetPublicKey(user_name))

		EncryptedQtable[state] = DecryptedQ_ct

		// if datas[0].V_t[state] == 1 {
		// 	fmt.Println(datas[0].Qvalue)
		// 	fmt.Println(Qold)
		// 	fmt.Println(DecryptedQ)
		// }
		// if datas[1].V_t[state] == 1 {
		// 	fmt.Println(datas[1].Qvalue)
		// 	fmt.Println(Qold)
		// 	fmt.Println(DecryptedQ)
		// }

		// EncryptedQtable[state] = testContext.Evaluator.AddNew(EncryptedQtable[i], re_fhe_v_and_w_Qnew)
		// EncryptedQtable[i] = testContext.Evaluator.SubNew(EncryptedQtable[i], re_fhe_v_and_w_Qold)
	}

	// fmt.Println("----------------")
	// fmt.Println(v)
	// fmt.Println()
	// fmt.Println(Qw)
	// fmt.Println()
	// fmt.Println(w)
}

func SecureQtableMultiUpdating(datas []utils.QvalueUpdateData, testContext *utils.TestParams, EncryptedQtable []*mkckks.Ciphertext, user_name string) {
	Nv := len(datas[0].V_t)
	Na := len(datas[0].W_t)

	all_Q_v_t := make([]float64, Nv)
	for state := 0; state < Nv; state++ {
		for user := 0; user < len(datas); user++ {
			if datas[user].V_t[state] == 1 {
				all_Q_v_t[state] += datas[user].Qvalue
			}
		}
	}

	/*
		行動(w_t)は行ベクトルのため、列ベクトルである状態(v_t)を行方向に拡張する
		v_t -> v_t_expanted
		[0, -> [0, 0, 0, 0]
		 :      :  :  :  :
		 0, -> [0, 0, 0, 0]
		 1, -> [1, 1, 1, 1]
		 0, -> [0, 0, 0, 0]
		 :      :  :  :  :
		 0] -> [0, 0, 0, 0]
	*/

	all_Q_v_t_expanded := make([]*mkckks.Ciphertext, Nv)
	for i := 0; i < Nv; i++ {
		if all_Q_v_t[i] == 0 {
			zeros := initializeZeros(Na, testContext.Params)
			all_Q_v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(zeros, testContext.PkSet.GetPublicKey(user_name))
		} else if all_Q_v_t[i] != 0 {
			ones := initializeQvalue(Na, all_Q_v_t[i], testContext.Params)
			all_Q_v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(ones, testContext.PkSet.GetPublicKey(user_name))
		}
	}

	all_v_t := make([]float64, Nv)
	for state := 0; state < Nv; state++ {
		for user := 0; user < len(datas); user++ {
			if datas[user].V_t[state] == 1 {
				all_v_t[state] = 1
			}
		}
	}

	all_v_t_expanded := make([]*mkckks.Ciphertext, Nv)
	for i := 0; i < Nv; i++ {
		if all_v_t[i] == 0 {
			zeros := initializeZeros(Na, testContext.Params)
			all_v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(zeros, testContext.PkSet.GetPublicKey(user_name))
		} else if all_v_t[i] == 1 {
			ones := initializeOnes(Na, testContext.Params)
			all_v_t_expanded[i] = testContext.Encryptor.EncryptMsgNew(ones, testContext.PkSet.GetPublicKey(user_name))
		}
	}

	// V_t[i] == 1 のときだけ w_t[i] = [0, 0, 1]になって，それ以外は w_t[i] = [0, 0, 0]になってほしい
	fhe_w_t := make([]*mkckks.Ciphertext, Nv)
	for i := 0; i < Nv; i++ {
		for user := 0; user < len(datas); user++ {
			if datas[user].V_t[i] == 1 {
				Was := initializeWa(Na, datas[user].W_t, testContext.Params)
				fhe_w_t[i] = testContext.Encryptor.EncryptMsgNew(Was, testContext.PkSet.GetPublicKey(user_name))

				if datas[0].V_t[i] == 1 && datas[1].V_t[i] == 1 {
					temp := make([]float64, Na)
					for j := 0; j < Na; j++ {
						temp[j] = datas[0].W_t[j] + datas[1].W_t[j]
					}

					Was := initializeWa(Na, temp, testContext.Params)
					fhe_w_t[i] = testContext.Encryptor.EncryptMsgNew(Was, testContext.PkSet.GetPublicKey(user_name))
				}

			} else {
				zeros := initializeZeros(Na, testContext.Params)
				fhe_w_t[i] = testContext.Encryptor.EncryptMsgNew(zeros, testContext.PkSet.GetPublicKey(user_name))
			}
		}

	}

	// Q_news_msg := mkckks.NewMessage(testContext.Params)
	// for i := 0; i < Na; i++ {
	// 	Q_news_msg.Value[i] = complex(Q_new, 0) // 虚部は0
	// }
	// fhe_Q_news := testContext.Encryptor.EncryptMsgNew(Q_news_msg, testContext.PkSet.GetPublicKey(user_name))

	for i := 0; i < Nv; i++ {
		// EncryptedQtable[i] = EncryptedQtable[i] + Qnew * v_t * w_t - Qold * v_t * w_t

		// calc: Qnew * (v_t * w_t)
		fhe_v_and_w_Qnew := testContext.Evaluator.MulRelinNew(all_Q_v_t_expanded[i], fhe_w_t[i], testContext.RlkSet)
		// fhe_v_and_w_Qnew = testContext.Evaluator.MulRelinNew(fhe_v_and_w_Qnew, fhe_Q_news, testContext.RlkSet)

		// calc: Qold * (v_t * w_t)
		fhe_v_and_w_Qold := testContext.Evaluator.MulRelinNew(all_v_t_expanded[i], fhe_w_t[i], testContext.RlkSet)
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

	/*
		行動(w_t)は行ベクトルのため、列ベクトルである状態(v_t)を行方向に拡張する
		v_t -> v_t_expanted
		[0, -> [0, 0, 0, 0]
		 :      :  :  :  :
		 0, -> [0, 0, 0, 0]
		 1, -> [1, 1, 1, 1]
		 0, -> [0, 0, 0, 0]
		 :      :  :  :  :
		 0] -> [0, 0, 0, 0]
	*/

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
		temp := testContext.Decryptor.Decrypt(EncryptedQtable[i], testContext.SkSet)
		EncryptedQtable[i] = testContext.Encryptor.EncryptMsgNew(temp, testContext.PkSet.GetPublicKey(user_name))

		// s_t[i] == 1: [1, ..., 1] * [Q1, ..., Qn] = [Q1, ..., Qn](s_t)
		// s_t[i] == 0: [0, ..., 0] * [Q1, ..., Qn] = [0 , ..., 0]
		v_t_expanded[i] = testContext.Evaluator.MulRelinNew(v_t_expanded[i], EncryptedQtable[i], testContext.RlkSet)

		// [0 , ..., 0] + ... + [Q1, ..., Qn] + ... + [0 , ..., 0] = [Q1, ..., Qn](s_t)
		actions = testContext.Evaluator.AddNew(actions, v_t_expanded[i])
	}

	return actions
}
