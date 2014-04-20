package libpaxos

import (
	"distributed2048/gameserver"
)

type proposal struct {
	Moves []gameserver.Move
}

func newProposal(moves []gameserver.Move) *proposal {
	return &proposal{moves}
}
