package lib2048

import (
	"distributed2048/rpc/paxosrpc"
)

const (
	BoardLen                 = 4
	FirstTileValue           = 2
	InitialTileCount         = 2
	InitialTileDoublePercent = 10
	EachTurnNewTileCount     = 1
)

type Game2048 interface {
	MakeMove(dir paxosrpc.Direction)
	GetScore() int
	GetBoard() [BoardLen][BoardLen]int
	IsGameOver() bool
	IsGameWon() bool
	String() string
}
