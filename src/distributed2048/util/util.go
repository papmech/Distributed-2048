package util

import (
	"distributed2048/rpc/paxosrpc"
	"io"
	"io/ioutil"
	"log"
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
