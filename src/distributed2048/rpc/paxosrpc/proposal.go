package paxosrpc

import (
	"distributed2048/lib2048"
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

type GameData struct {
	Grid        lib2048.Grid
	Score       int
	RandCurrent uint32
}

func NewGameData(game lib2048.Game2048) *GameData {
	return &GameData{game.GetBoard(), game.GetScore(), game.GetRand().GetCurrent()}
}

func (gd *GameData) CopyInto(game lib2048.Game2048) {
	game.SetGrid(gd.Grid)
	game.SetScore(gd.Score)
	game.GetRand().SetCurrent(gd.RandCurrent)
}

type ProposalValue struct {
	Moves []lib2048.Move
	Game  GameData
}

type Proposal struct {
	Number            ProposalNumber
	CommandSlotNumber uint32
	Value             ProposalValue
}

func NewProposal(number, commandSlotNumber, nodeID uint32, value ProposalValue) *Proposal {
	return &Proposal{
		ProposalNumber{number, nodeID},
		commandSlotNumber,
		value,
	}
}
