package utils

import (
	"MKpprlgoFrozenLake/mkckks"
	"MKpprlgoFrozenLake/mkrlwe"
	"fmt"

	"github.com/ldsec/lattigo/v2/ckks"
	"github.com/ldsec/lattigo/v2/ring"
	"github.com/ldsec/lattigo/v2/rlwe"
	"github.com/ldsec/lattigo/v2/utils"
)

// ckks parameters
var (
	PN15QP880 = ckks.ParametersLiteral{
		LogN:     15,
		LogSlots: 2, // default: 14
		//60 + 13x54
		Q: []uint64{
			0xfffffffff6a0001,

			0x3fffffffd60001, 0x3fffffffca0001,
			0x3fffffff6d0001, 0x3fffffff5d0001,
			0x3fffffff550001, 0x3fffffff390001,
			0x3fffffff360001, 0x3fffffff2a0001,
			0x3fffffff000001, 0x3ffffffefa0001,
			0x3ffffffef40001, 0x3ffffffed70001,
			0x3ffffffed30001,
		},
		P: []uint64{
			//59 x 2
			0x7ffffffffe70001, 0x7ffffffffe10001,
		},
		Scale: 1 << 54,
		Sigma: rlwe.DefaultSigma,
	}
	PN14QP439 = ckks.ParametersLiteral{
		LogN:     14,
		LogSlots: 13,
		Q: []uint64{
			// 59 + 5x52
			0x7ffffffffe70001,

			0xffffffff00001, 0xfffffffe40001,
			0xfffffffe20001, 0xfffffffbe0001,
			0xfffffffa60001,
		},
		P: []uint64{
			// 60 x 2
			0xffffffffffc0001, 0xfffffffff840001,
		},
		Scale: 1 << 52,
		Sigma: rlwe.DefaultSigma,
	}

	FAST_BUT_NOT_128 = ckks.ParametersLiteral{
		LogN:     7,
		LogQ:     []int{35, 60, 60},
		LogP:     []int{45, 45},
		LogSlots: 2,
		Scale:    1 << 10, // 30
	}

	PPRL_PARAMS = ckks.ParametersLiteral{
		LogN:     13,                // 13
		LogQ:     []int{35, 60, 60}, // []int{55, 40, 40},
		LogP:     []int{45, 45},
		LogSlots: 2,
		Scale:    1 << 30,
	}
)

type TestParams struct {
	Params mkckks.Parameters
	RingQ  *ring.Ring
	RingP  *ring.Ring
	Prng   utils.PRNG
	Kgen   *mkrlwe.KeyGenerator
	SkSet  *mkrlwe.SecretKeySet
	PkSet  *mkrlwe.PublicKeySet
	RlkSet *mkrlwe.RelinearizationKeySet
	RtkSet *mkrlwe.RotationKeySet
	CjkSet *mkrlwe.ConjugationKeySet

	Encryptor *mkckks.Encryptor
	Decryptor *mkckks.Decryptor
	Evaluator *mkckks.Evaluator
	Idset     *mkrlwe.IDSet
}

func GenTestParams(defaultParam mkckks.Parameters, idset *mkrlwe.IDSet) (testContext *TestParams, err error) {

	testContext = new(TestParams)

	testContext.Params = defaultParam

	testContext.Kgen = mkckks.NewKeyGenerator(testContext.Params)

	testContext.SkSet = mkrlwe.NewSecretKeySet()
	testContext.PkSet = mkrlwe.NewPublicKeyKeySet()
	testContext.RlkSet = mkrlwe.NewRelinearizationKeyKeySet(defaultParam.Parameters)
	testContext.RtkSet = mkrlwe.NewRotationKeySet()
	testContext.CjkSet = mkrlwe.NewConjugationKeySet()

	// gen sk, pk, rlk, rk

	for id := range idset.Value {
		sk, pk := testContext.Kgen.GenKeyPair(id)
		r := testContext.Kgen.GenSecretKey(id)
		rlk := testContext.Kgen.GenRelinearizationKey(sk, r)
		testContext.SkSet.AddSecretKey(sk)
		testContext.PkSet.AddPublicKey(pk)
		testContext.RlkSet.AddRelinearizationKey(rlk)
	}

	testContext.RingQ = defaultParam.RingQ()

	if testContext.Prng, err = utils.NewPRNG(); err != nil {
		return nil, err
	}

	testContext.Encryptor = mkckks.NewEncryptor(testContext.Params)
	testContext.Decryptor = mkckks.NewDecryptor(testContext.Params)

	testContext.Evaluator = mkckks.NewEvaluator(testContext.Params)

	return testContext, nil

}

func GeneratePlaintextAndCiphertext(testContext *TestParams, id string, a, b complex128) (msg *mkckks.Message, ciphertext *mkckks.Ciphertext) {

	Params := testContext.Params
	logSlots := testContext.Params.LogSlots()

	msg = mkckks.NewMessage(Params)

	for i := 0; i < 1<<logSlots; i++ {
		msg.Value[i] = complex(utils.RandFloat64(real(a), real(b)), utils.RandFloat64(imag(a), imag(b)))
	}

	if testContext.Encryptor != nil {
		ciphertext = testContext.Encryptor.EncryptMsgNew(msg, testContext.PkSet.GetPublicKey(id))
	} else {
		panic("cannot newTestVectors: Encryptor is not initialized!")
	}

	return msg, ciphertext
}

func testEncAndDec(testContext *TestParams, userList []string) {
	numUsers := len(userList)
	msgList := make([]*mkckks.Message, numUsers)
	ctList := make([]*mkckks.Ciphertext, numUsers)

	SkSet := testContext.SkSet
	dec := testContext.Decryptor

	for i := range userList {
		msgList[i], ctList[i] = GeneratePlaintextAndCiphertext(testContext, userList[i], complex(-1, 0), complex(1, 0))
	}

	user1 := "user1"
	user2 := "user2"
	idset1 := mkrlwe.NewIDSet()
	idset2 := mkrlwe.NewIDSet()
	idset1.Add(user1)
	idset2.Add(user2)

	ct3 := testContext.Evaluator.AddNew(ctList[0], ctList[1])
	ct4 := testContext.Evaluator.MulRelinNew(ctList[0], ctList[1], testContext.RlkSet)

	//testContext.Evaluator.MultByConst(ct3, constant, ct3)
	//ct3.Scale *= float64(constant)
	//testContext.Evaluator.Rescale(ct3, Params.Scale(), ct3)
	msg3Out := testContext.Decryptor.Decrypt(ct3, testContext.SkSet)
	msg4Out := testContext.Decryptor.Decrypt(ct4, testContext.SkSet)

	fmt.Println("Enc and Dec without any calculation")
	for i := range userList {
		msgOut := dec.Decrypt(ctList[i], SkSet)
		fmt.Printf("user-%d:\nplaintext: %g,\ndecrypted: %g\n", i, msgList[i], msgOut)
	}

	fmt.Println("Add: user1 + user2")
	fmt.Println(msg3Out)
	fmt.Println("Mul: user1 * user2")
	fmt.Println(msg4Out)
}
