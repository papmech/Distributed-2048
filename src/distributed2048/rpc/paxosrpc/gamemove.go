package paxosrpc

import (
	"fmt"
	"time"
)

type Direction int

const layout = "15:04:05.12"
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

func (m *Move) String() string {
	var moveString string
	switch m.Direction {
	case Up:
		moveString = "Up"
	case Down:
		moveString = "Down"
	case Left:
		moveString = "Left"
	case Right:
		moveString = "Right"
	}

	return fmt.Sprintf("'%s': %s", m.Time.Format(layout), moveString)
}
