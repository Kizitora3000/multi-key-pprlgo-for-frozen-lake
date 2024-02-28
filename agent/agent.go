package agent

import (
	"MKpprlgoFrozenLake/environment"
	"MKpprlgoFrozenLake/mkckks"
	"MKpprlgoFrozenLake/position"
	"MKpprlgoFrozenLake/pprl"
	"MKpprlgoFrozenLake/utils"
	"fmt"
	"math"
	"math/rand"
)

type Agent struct {
	Env        *environment.Environment
	actionNum  int
	stateNum   int
	lakeHeight int
	lakeWidth  int
	InitValQ   float64
	Epsilon    float64
	Alpha      float64
	Gamma      float64
	Qtable     [][]float64 // Qテーブルの状態は1次元とする (状態をposition.Positionにすると暗号化時に処理できない)
}

const (
	INITIAL_VAL_Q = 0
	EPSILON       = 0.1
	ALPHA         = 0.1
	GAMMA         = 0.9
)

func NewAgent(env *environment.Environment) *Agent {
	actionNum := len(env.ActionSpace)
	stateNum := env.Height() * env.Width()
	lakeHeight := env.Height()
	lakeWidth := env.Width()

	// Qtable[stateNum][actionNum]の二次元配列を作成してInitValQで初期化
	Qtable := make([][]float64, stateNum)
	for i := range Qtable {
		Qtable[i] = make([]float64, actionNum)
		for j := range Qtable[i] {
			Qtable[i][j] = INITIAL_VAL_Q
		}
	}

	return &Agent{
		Env:        env,
		actionNum:  actionNum,
		stateNum:   stateNum,
		lakeHeight: lakeHeight,
		lakeWidth:  lakeWidth,
		InitValQ:   INITIAL_VAL_Q,
		Epsilon:    EPSILON,
		Alpha:      ALPHA,
		Gamma:      GAMMA,
		Qtable:     Qtable,
	}
}

func (a *Agent) QtableReset(env *environment.Environment) {
	// Qtable[stateNum][actionNum]の二次元配列を作成してInitValQで初期化
	for i := range a.Qtable {
		a.Qtable[i] = make([]float64, a.actionNum)
		for j := range a.Qtable[i] {
			a.Qtable[i][j] = INITIAL_VAL_Q
		}
	}
}

func (e *Agent) Learn(state position.Position, act int, rwd int, next_state position.Position, testContext *utils.TestParams, encryptedQtable []*mkckks.Ciphertext, user_name string) {
	state_1D := e.convert2DTo1D(state)
	next_state_1D := e.convert2DTo1D(next_state)

	target := float64(0)
	target = float64(rwd) + e.Gamma*e.maxValue(e.Qtable[next_state_1D]) // rwdは整数値なので実数値にキャストする

	e.Qtable[state_1D][act] = (1-e.Alpha)*e.Qtable[state_1D][act] + e.Alpha*target

	v_t := make([]float64, e.stateNum)
	w_t := make([]float64, e.actionNum)
	v_t[state_1D] = 1
	w_t[act] = 1

	Qnew := e.Qtable[state_1D][act]
	pprl.SecureQtableUpdating(v_t, w_t, Qnew, e.stateNum, e.actionNum, testContext, encryptedQtable, user_name)
}

func (e *Agent) Trajectory(state position.Position, act int, rwd int, next_state position.Position, testContext *utils.TestParams, encryptedQtable []*mkckks.Ciphertext) ([]float64, []float64, float64) {
	state_1D := e.convert2DTo1D(state)
	next_state_1D := e.convert2DTo1D(next_state)

	target := float64(0)
	target = float64(rwd) + e.Gamma*e.maxValue(e.Qtable[next_state_1D]) // rwdは整数値なので実数値にキャストする

	e.Qtable[state_1D][act] = (1-e.Alpha)*e.Qtable[state_1D][act] + e.Alpha*target

	v_t := make([]float64, e.stateNum)
	w_t := make([]float64, e.actionNum)
	v_t[state_1D] = 1
	w_t[act] = 1

	Qnew := e.Qtable[state_1D][act]

	return v_t, w_t, Qnew
}

func (e *Agent) maxValue(slice []float64) float64 {
	maxValue := slice[0]
	for _, v := range slice {
		if v > maxValue {
			maxValue = v
		}
	}
	return maxValue
}

// Qテーブルの状態(1次元)に格納するため，二次元座標を一次元インデックスに変換
func (e *Agent) convert2DTo1D(state position.Position) int {
	return state.Y*e.lakeWidth + state.X
}

// ランダムに行動を選択
func (a *Agent) ChooseRandomAction() int {
	return rand.Intn(a.actionNum) // 0からactionNum-1までの範囲でランダムに整数を返す
}

// εグリーディー方策
func (a *Agent) EpsilonGreedyAction(state position.Position) int {
	state_1D := a.convert2DTo1D(state)

	// εより小さいランダムな値を生成してランダムに行動を選択
	if rand.Float64() < a.Epsilon {
		return a.ChooseRandomAction()
	}

	// 最大のQ値を持つ行動を選択
	maxAction := 0
	maxQValue := a.Qtable[state_1D][0]
	for action, qValue := range a.Qtable[state_1D] {
		if qValue > maxQValue {
			maxAction = action
			maxQValue = qValue
		}
	}

	return maxAction
}

// 実数を指定された桁数で切り捨てる
func TruncateFloat(value float64, places int) float64 {
	if math.Abs(value) < float64(math.Pow(10, float64(-places))) {
		return 0.0
	}

	shift := math.Pow(10, float64(places))
	return math.Floor(value*shift) / shift
}

// 複素数の実部と虚部を指定された桁数で切り捨てる
func TruncateComplex(c complex128) complex128 {
	const places = 5 // 小数点第4位
	realPart := real(c)
	imagPart := imag(c)

	truncatedReal := TruncateFloat(realPart, places)
	truncatedImag := TruncateFloat(imagPart, places)

	return complex(truncatedReal, truncatedImag)
}

// εグリーディー方策(クラウド上のQテーブルから選択)
func (a *Agent) SecureEpsilonGreedyAction(state position.Position, testContext *utils.TestParams, encryptedQtable []*mkckks.Ciphertext, user_name string) int {
	// εより小さいランダムな値を生成してランダムに行動を選択
	if rand.Float64() < a.Epsilon {
		return a.ChooseRandomAction()
	}

	state_1D := a.convert2DTo1D(state)
	v_t := make([]float64, a.stateNum)
	v_t[state_1D] = 1

	// 最大のQ値を持つ行動を選択
	actions_Q_in_state := pprl.SecureActionSelection(v_t, a.stateNum, a.actionNum, testContext, encryptedQtable, user_name)
	actions_Q_in_state_msg := testContext.Decryptor.Decrypt(actions_Q_in_state, testContext.SkSet)

	// fmt.Println(actions_Q_in_state_msg.Value)
	// 復号時に生じる極小な誤差は切り捨てる
	for i := 0; i < a.actionNum; i++ {
		actions_Q_in_state_msg.Value[i] = TruncateComplex(actions_Q_in_state_msg.Value[i])
	}
	// fmt.Println(actions_Q_in_state_msg.Value)

	maxAction := 0
	maxQValue := real(actions_Q_in_state_msg.Value[0]) // 実部だけ抽出

	for idx := 0; idx < a.actionNum; idx++ {
		qValue := real(actions_Q_in_state_msg.Value[idx]) // 実部だけ抽出
		if qValue > maxQValue {
			maxAction = idx
			maxQValue = qValue
		}
	}

	return maxAction
}

// 貪欲方策
func (a *Agent) GreedyAction(state position.Position) int {
	state_1D := a.convert2DTo1D(state)

	// 最大のQ値を持つ行動を選択
	maxAction := 0
	maxQValue := a.Qtable[state_1D][0]
	for action, qValue := range a.Qtable[state_1D] {
		if qValue > maxQValue {
			maxAction = action
			maxQValue = qValue
		}
	}

	return maxAction
}

func (a *Agent) ShowQTable() {
	// 行動インデックスに対応する方向の文字列
	actionSymbols := map[int]string{
		0: "↑",
		1: "↓",
		2: "←",
		3: "→",
	}

	fmt.Println("Qtable:")

	for stateIndex, actions := range a.Qtable {
		// 状態を二次元座標に変換して表示
		stateY := stateIndex / a.lakeWidth
		stateX := stateIndex % a.lakeWidth
		fmt.Printf("State [Y: %d, X: %d]: ", stateY, stateX)

		for actionIndex, qValue := range actions {
			// actionIndex を方向の文字列に変換して表示
			actionSymbol := actionSymbols[actionIndex]
			fmt.Printf("%s: %.2f ", actionSymbol, qValue)
		}
		fmt.Println()
	}
}

func (a *Agent) ShowOptimalPath(env *environment.Environment) {
	currentState := env.Reset() // 環境をリセットしてスタート位置を取得
	fmt.Println("Optimal Path: ")

	// 行動インデックスに対応する方向の文字列
	actionSymbols := map[int]string{
		0: "↑",
		1: "↓",
		2: "←",
		3: "→",
	}

	for {
		action := a.GreedyAction(currentState)

		// 最適な行動に基づいて次の状態を決定
		nextState := env.NextState(currentState, action)

		// 経路を出力
		if currentState == env.StartPos {
			fmt.Println("START")
		}
		fmt.Printf("state: %s,  action: %s\n", currentState, actionSymbols[action])
		currentState = nextState

		if currentState == env.GoalPos {
			fmt.Println("GOAL")
			break // ゴールに到達したらループを終了
		}
	}
}

func (e *Agent) GetActionNum() int {
	return e.actionNum
}

func (e *Agent) GetStateNum() int {
	return e.stateNum
}
