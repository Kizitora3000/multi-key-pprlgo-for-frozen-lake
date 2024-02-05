package main

import (
	"MKpprlgoFrozenLake/frozenlake"
	"fmt"
)

func main() {
	lake := frozenlake.FrozenLake3x3
	fmt.Println(lake.LakeMap)
}
