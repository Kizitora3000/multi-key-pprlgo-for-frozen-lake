package environment

import (
	"MKpprlgoFrozenLake/frozenlake"
	"MKpprlgoFrozenLake/position"
)

const GOAL_REWARD = 1000

type Environment struct {
	frozenLake  frozenlake.FrozenLake
	actionSpace []int             // エージェントの行動空間
	agentState  position.Position // エージェントの現在位置
	rewards     [][]int
	isHole      map[position.Position]bool // True: 穴, False: 地面
}

func NewEnvironment(lake frozenlake.FrozenLake) *Environment {
	frozenLake := lake
	actionSpace := []int{0, 1, 2, 3}  // 0: "↑", 1: "↓", 2: "←", 3: "→"
	agentState := frozenLake.StartPos // エージェントの位置はスタート地点で初期化
	isHole := make(map[position.Position]bool)

	// Height x Width の2次元配列を作成
	rewards := make([][]int, frozenLake.Height)
	for i := range rewards {
		rewards[i] = make([]int, frozenLake.Width)
	}

	// 報酬を設定する: "o"(地面) → 0, "x"(穴) → -1
	for y, row := range frozenLake.LakeMap {
		for x, cell := range row {
			switch cell {
			case "o": // 地面
				rewards[y][x] = 0
				isHole[position.Position{Y: y, X: x}] = false
			case "x": // 穴
				rewards[y][x] = -1
				isHole[position.Position{Y: y, X: x}] = true
			}
		}
	}

	// ゴール地点の報酬を設定
	rewards[frozenLake.GoalPos.Y][frozenLake.GoalPos.X] = GOAL_REWARD

	return &Environment{
		frozenLake:  frozenLake,
		actionSpace: actionSpace,
		agentState:  agentState,
		rewards:     rewards,
		isHole:      isHole,
	}
}

func (e *Environment) Height() int {
	return e.frozenLake.Height
}

func (e *Environment) Width() int {
	return e.frozenLake.Width
}

func (e *Environment) Reward(nextState position.Position) int {
	return e.rewards[nextState.Y][nextState.X]
}

func (e *Environment) Reset() position.Position {
	e.agentState = e.frozenLake.StartPos
	return e.agentState
}

func (e *Environment) NextState(state position.Position, action int) position.Position {
	// 行動空間(0: "↑", 1: "↓", 2: "←", 3: "→")に基づいて移動方向を設定
	actionMoveMap := []position.Position{{Y: 1, X: 0}, {Y: -1, X: 0}, {Y: 0, X: -1}, {Y: 0, X: 1}}
	move := actionMoveMap[action]

	// 現在の状態(state) + 移動方向(move) = 次の状態(nextState)
	nextState := position.Position{Y: state.Y + move.Y, X: state.X + move.X}

	// 移動先が画面外の場合は移動しない
	if nextState.X < 0 || nextState.X >= e.Width() || nextState.Y < 0 || nextState.Y >= e.Height() {
		nextState = state
	}

	return nextState
}
