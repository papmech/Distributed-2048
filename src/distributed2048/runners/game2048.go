package main

import (
	"distributed2048/lib2048"
	"distributed2048/rpc/paxosrpc"
	"fmt"
)

func main() {
	game, _ := lib2048.NewGame2048()

	fmt.Println(game.String())

	for !game.IsGameOver() {
		var input string
		fmt.Scanf("%s", &input)
		var dir paxosrpc.Direction
		switch input {
		case "w":
			dir = paxosrpc.Up
		case "a":
			dir = paxosrpc.Left
		case "s":
			dir = paxosrpc.Down
		case "d":
			dir = paxosrpc.Right
		}
		game.MakeMove(dir)
		fmt.Println(game.String())
	}
}
