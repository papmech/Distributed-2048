package main

import (
	"distributed2048/lib2048"
	"fmt"
)

func main() {
	game := lib2048.NewGame2048()

	r := game.GetRand()
	for i := 0; i < 100; i++ {
		fmt.Println(r.Int() % 4)
	}

	fmt.Println(game.String())

	for !game.IsGameOver() {
		var input string
		fmt.Scanf("%s", &input)
		var dir lib2048.Direction
		switch input {
		case "w":
			dir = lib2048.Up
		case "a":
			dir = lib2048.Left
		case "s":
			dir = lib2048.Down
		case "d":
			dir = lib2048.Right
		}
		game.MakeMove(dir)
		fmt.Println(game.String())
	}
}
