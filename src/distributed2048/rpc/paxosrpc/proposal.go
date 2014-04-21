package paxosrpc

import (
	"fmt"
)

type ProposalNumber struct {
	Number uint32
	NodeID uint32
}

func (a *ProposalNumber) LessThan(b *ProposalNumber) bool {
	return a.Number < b.Number || (a.Number == b.Number && a.NodeID < b.NodeID)
}

func (a *ProposalNumber) Equal(b *ProposalNumber) bool {
	return a.NodeID == b.NodeID && a.Number == b.Number
}

func (a *ProposalNumber) GreaterThan(b *ProposalNumber) bool {
	return a.Number > b.Number || (a.Number == b.Number && a.NodeID > b.NodeID)
}

func (a *ProposalNumber) String() string {
	return fmt.Sprintf("(%d, %d)", a.Number, a.NodeID)
}

type Proposal struct {
	Number            ProposalNumber
	CommandSlotNumber uint32
	Value             []Move
}

func NewProposal(number, commandSlotNumber, nodeID uint32, moves []Move) *Proposal {
	return &Proposal{
		ProposalNumber{number, nodeID},
		commandSlotNumber,
		moves,
	}
}
