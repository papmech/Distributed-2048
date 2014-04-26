package tests
import (
	"distributed2048/lib2048"
	"fmt"
//	"strconv"
)

type Gamecomparison struct {
	Board lib2048.Grid
	Score int
	Over bool
	Won bool
	Cboard lib2048.Grid
	Cscore int
	Cover bool
	Cwon bool
}

func (g *Gamecomparison) CompareGame() bool {
//	if g.Score != g.Cscore {
//		fmt.Println("Expected score = " + strconv.Itoa(g.Score) + " Got: " + strconv.Itoa(g.Cscore))
//		return false
//	}
//	if g.Over != g.Cover {
//		if g.Over {
//			fmt.Println("Expected game state is game over, got game not over")
//		} else {
//			fmt.Println("Expected game state is game not over, got game over")
//		}
//		return false
//	}
//	if g.Won != g.Cwon {
//		if g.Won {
//			fmt.Println("Expected game state is game won, got game not won")
//		} else {
//			fmt.Println("Expected game state is game not won, got game won")
//		}
//		return false
//	}
	if g.Board != g.Cboard {
		fmt.Print("expected board is ")
		fmt.Println(g.Board)
		fmt.Print("got board ")
		fmt.Println(g.Cboard)
		return false
	}
	return true
}
