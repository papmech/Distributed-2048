package paxosrpc

import (
	"distributed2048/gameserver"
)

type ProposalNumber struct {
	Number uint32
	NodeID uint32
}

func (a *ProposalNumber) LessThan(b ProposalNumber) bool {
	return a.Number < b.Number || (a.Number == b.Number && a.NodeID < b.NodeID)
}

func (a *ProposalNumber) Equal(b ProposalNumber) bool {
	return a.NodeID == b.NodeID && a.Number == b.Number
}

func (a *ProposalNumber) GreaterThan(b ProposalNumber) bool {
	return a.Number > b.Number || (a.Number == b.Number && a.NodeID > b.NodeID)
}

type Proposal struct {
	Number ProposalNumber
	Value  []gameserver.Move
}

func NewProposal(number, nodeID uint32, moves []gameserver.Move) *Proposal {
	return &Proposal{
		ProposalNumber{number, nodeID},
		moves,
	}
}
