package util

import (
	"distributed2048/rpc/paxosrpc"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"distributed2048/lib2048"
)

const (
	CENTRALPORT = 25340
	CENTRALHOSTPOST = "localhost:25340"
	GAMESERVERPORT = 15551
	LOCALHOST = "localhost"
	DEFAULTINTERVAL = 5
	DEFAULTPATTERN = "/"
	CSFAIL = "PHAIL: COULD NOT START CENTRAL SERVER"
	GSFAIL = "PHAIL: COULD NOT START GAME SERVER"
	CFAIL = "PHAIL: COULD NOT START CLIENT"
)

func NewLogger(enabled bool, prefix string, out io.Writer) *log.Logger {
	w := ioutil.Discard
	if enabled {
		w = out
	}
	return log.New(w, "["+prefix+"] ", log.Lshortfile)
}

func MovesString(moves []paxosrpc.Move) string {
	result := "{\n"
	for _, move := range moves {
		result += "    " + move.String() + "\n"
	}
	result += "}"
	return result
}

func RandomMove() *paxosrpc.Move {
	num := rand.Int() % 4
	var dir paxosrpc.Direction
	switch num {
	case 0:
		dir = paxosrpc.Up
	case 1:
		dir = paxosrpc.Left
	case 2:
		dir = paxosrpc.Down
	case 3:
		dir = paxosrpc.Right
	}
	return paxosrpc.NewMove(dir)
}

func CompareDir(dir1, dir2 paxosrpc.Direction) bool {
	return dir1 > dir2
}

func CalculateGameState(initial lib2048.Grid, moves []int) lib2048.Grid {
	game := lib2048.NewGame2048()
	game.SetGameState(initial)

}
