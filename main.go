package main

import (
	"MKpprlgoFrozenLake/environment"
	"MKpprlgoFrozenLake/frozenlake"
	"fmt"
)

const EPISODES = 1

func main() {
	lake := frozenlake.FrozenLake3x3
	env := environment.NewEnvironment(lake)

	state := env.Reset()
	nextState, reward, done := env.Step(3)

	fmt.Println(state, nextState, reward, done)
}
