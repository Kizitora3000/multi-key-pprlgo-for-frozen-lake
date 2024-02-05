package frozenlake

import "MKpprlgoFrozenLake/position"

type FrozenLake struct {
	Width    int               // 湖の幅
	Height   int               // 湖の高さ
	LakeMap  [][]string        // 湖の状態 ("o": 地面, "x": 穴)
	StartPos position.Position // スタート地点
	GoalPos  position.Position // ゴール地点
}

var (
	FrozenLake3x3 = FrozenLake{
		Width:  3,
		Height: 3,
		LakeMap: [][]string{
			{"o", "x", "x"},
			{"o", "o", "o"},
			{"x", "x", "o"},
		},
		StartPos: position.Position{
			X: 0,
			Y: 0,
		},
		GoalPos: position.Position{
			X: 2, // Width - 1
			Y: 2, // Height - 1
		},
	}

	FrozenLake4x4 = FrozenLake{
		Width:  4,
		Height: 4,
		LakeMap: [][]string{
			{"o", "o", "x", "x"},
			{"o", "x", "o", "x"},
			{"o", "o", "o", "o"},
			{"o", "x", "x", "o"},
		},
		StartPos: position.Position{
			X: 0,
			Y: 0,
		},
		GoalPos: position.Position{
			X: 3, // Width - 1
			Y: 3, // Height - 1
		},
	}
)
