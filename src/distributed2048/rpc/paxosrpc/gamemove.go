package paxosrpc

import (
	"time"
)

type Direction int

const (
	Up Direction = iota + 1
	Left
	Down
	Right
)

type Move struct {
	Time      time.Time
	Direction Direction
}

func NewMove(dir Direction) *Move {
	return &Move{time.Now(), dir}
}
