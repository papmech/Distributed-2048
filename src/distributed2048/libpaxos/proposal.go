package libpaxos

import (
	"distributed2048/gameserver"
)

type Proposal struct {
	Moves []gameserver.Move
}

func NewProposal(moves []gameserver.Move) *Proposal {
	return &Proposal{moves}
}
