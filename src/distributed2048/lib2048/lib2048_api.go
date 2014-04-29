// ported from https://github.com/olafurw/ML2048

package lib2048

import (
	"distributed2048/libsimplerand"
)

const (
	BoardLen                 = 4
	FirstTileValue           = 2
	InitialTileCount         = 2
	InitialTileDoublePercent = 10
	EachTurnNewTileCount     = 1
)

type Game2048 interface {
	MakeMove(dir Direction)
	GetScore() int
	GetBoard() Grid
	GetRand() *libsimplerand.SimpleRand
	IsGameOver() bool
	IsGameWon() bool
	String() string
	Equals(game Game2048) bool
	SetGrid(grid Grid)
	SetScore(score int)
	CloneFrom(game Game2048)
}
