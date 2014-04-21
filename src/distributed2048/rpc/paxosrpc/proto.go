package paxosrpc

type Status int

type Node struct {
	ID       uint32
	HostPort string
}

const (
	OK Status = iota + 1
	Reject
	DecidedValueExists
)

type ReceivePrepareArgs struct {
	Node              Node
	ProposalNumber    ProposalNumber
	CommandSlotNumber uint32
}

type ReceivePrepareReply struct {
	Status              Status
	HasAcceptedProposal bool
	AcceptedProposal    Proposal
	DecidedSlotNumber   uint32
	DecidedValue        []Move
}

type ReceiveAcceptArgs struct {
	Proposal Proposal
}

type ReceiveAcceptReply struct {
	Status
}

type ReceiveDecideArgs struct {
	Proposal Proposal
}

type ReceiveDecideReply struct {
}
