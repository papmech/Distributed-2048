package gameserver

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
