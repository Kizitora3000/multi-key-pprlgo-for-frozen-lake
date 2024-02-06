package main

import (
	"MKpprlgoFrozenLake/agent"
	"MKpprlgoFrozenLake/environment"
	"MKpprlgoFrozenLake/frozenlake"
)

const EPISODES = 1000000

func main() {
	lake := frozenlake.FrozenLake3x3
	env := environment.NewEnvironment(lake)
	agt := agent.NewAgent(env)

	for episode := 0; episode < EPISODES; episode++ {
		state := env.Reset()
		for {
			action := agt.ChooseRandomAction()
			next_state, reward, done := env.Step(action)
			agt.Learn(state, action, reward, next_state)

			if done {
				break
			}
			state = next_state

		}
	}

	agt.DisplayQTable()
	agt.DisplayOptimalPath(env)
}
