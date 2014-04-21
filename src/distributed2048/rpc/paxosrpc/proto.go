package paxosrpc

type Status int

type Node struct {
	ID       uint32
	HostPort string
}

const (
	OK Status = iota + 1
	Reject
)

type ReceivePrepareArgs struct {
	Node     Node
	Proposal Proposal
}

type ReceivePrepareReply struct {
	Status              Status
	HasAcceptedProposal bool
	AcceptedProposal    Proposal
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
