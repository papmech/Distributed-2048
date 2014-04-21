package util

import (
	"distributed2048/rpc/paxosrpc"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
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
