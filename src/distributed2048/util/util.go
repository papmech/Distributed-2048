package util

import (
	"distributed2048/lib2048"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"time"
)

const (
	CENTRALPORT     = 25340
	CENTRALHOSTPOST = "localhost:25340"
	GAMESERVERPORT  = 15551
	LOCALHOST       = "localhost"
	DEFAULTINTERVAL = 500
	DEFAULTPATTERN  = "/"
	CSFAIL          = "PHAIL: COULD NOT START CENTRAL SERVER"
	GSFAIL          = "PHAIL: COULD NOT START GAME SERVER"
	CFAIL           = "PHAIL: COULD NOT START CLIENT"
	GAMESTATEERR    = "PHAIL: Not able to get game state from client"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

type ClientMove struct {
	Direction int
}

type Game2048State struct {
	Won   bool
	Over  bool
	Grid  lib2048.Grid
	Score int
}

func (s *Game2048State) String() string {
	game := lib2048.NewGame2048()
	game.SetGrid(s.Grid)
	game.SetScore(s.Score)
	return fmt.Sprintf("%sWon: %s\nOver: %s\n", game.String(), s.Won, s.Over)
}

func NewLogger(enabled bool, prefix string, out io.Writer) *log.Logger {
	w := ioutil.Discard
	if enabled {
		w = out
	}
	return log.New(w, "["+prefix+"] ", log.Lshortfile)
}

func MovesString(moves []lib2048.Move) string {
	result := "{\n"
	for _, move := range moves {
		result += "    " + move.String() + "\n"
	}
	result += "}"
	return result
}

func RandomMove() *lib2048.Move {
	num := r.Int() % 4
	var dir lib2048.Direction
	switch num {
	case 0:
		dir = lib2048.Up
	case 1:
		dir = lib2048.Left
	case 2:
		dir = lib2048.Down
	case 3:
		dir = lib2048.Right
	}
	return lib2048.NewMove(dir)
}

func CompareDir(dir1, dir2 lib2048.Direction) bool {
	return dir1 > dir2
}

func CalculateGameState(initial lib2048.Game2048, moves []*lib2048.Move) lib2048.Game2048 {
	game := lib2048.NewGame2048()
	game.CloneFrom(initial)
	for _, m := range moves {
		game.MakeMove(m.Direction)
	}
	return game
}
