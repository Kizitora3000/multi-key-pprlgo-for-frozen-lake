package agent

import (
	"MKpprlgoFrozenLake/environment"
	"MKpprlgoFrozenLake/position"
	"fmt"
	"math/rand"
	"time"
)

type Agent struct {
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

func (e *Agent) Learn(state position.Position, act int, rwd int, next_state position.Position) {
	state_1D := e.convert2DTo1D(state)
	next_state_1D := e.convert2DTo1D(next_state)

	target := float64(0)
	target = float64(rwd) + e.Gamma*e.maxValue(e.Qtable[next_state_1D]) // rwdは整数値なので実数値にキャストする

	e.Qtable[state_1D][act] = (1-e.Alpha)*e.Qtable[state_1D][act] + e.Alpha*target

	// Qnew := e.Qtable[state_1D][act]
	v_t := make([]float64, e.stateNum) // マジックナンバー とりあえずUCIのデータセットの血糖値は最大で501
	w_t := make([]float64, e.actionNum)
	v_t[state_1D] = 1
	w_t[act] = 1
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
	rand.Seed(time.Now().UnixNano()) // ランダムなシード値で初期化
	return rand.Intn(a.actionNum)    // 0からactionNum-1までの範囲でランダムに整数を返す
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
		fmt.Printf("State (%d, %d): ", stateY, stateX)

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
		currentState1D := a.convert2DTo1D(currentState)
		bestAction := 0
		bestQValue := a.Qtable[currentState1D][0]

		// 最適な行動（Q値が最大の行動）を選択
		for action, qValue := range a.Qtable[currentState1D] {
			if qValue > bestQValue {
				bestAction = action
				bestQValue = qValue
			}
		}

		// 最適な行動に基づいて次の状態を決定
		nextState := env.NextState(currentState, bestAction)

		// 経路を出力
		if currentState == env.StartPos {
			fmt.Println("START")
		}
		fmt.Printf("state: %s,  action: %s\n", currentState, actionSymbols[bestAction])
		currentState = nextState

		if currentState == env.GoalPos {
			fmt.Println("GOAL")
			break // ゴールに到達したらループを終了
		}
	}
}
