package tests
import (
	"distributed2048/lib2048"
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
	if g.Score != g.Cscore {
		return false
	}
	if g.Over != g.Cover {
		return false
	}
	if g.Won != g.Cwon {
		return false
	}
	if g.Board != g.Cboard {
		return false
	}
	return true
}
