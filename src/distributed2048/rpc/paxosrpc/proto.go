package paxosrpc

type Status int

const (
	OK Status = iota + 1
	Accept
	Reject
)

type Node struct {
	ID       uint32
	HostPort string
}

type ReceivePrepareArgs struct {
	Node     Node
	Proposal Proposal
}

type ReceivePrepareReply struct {
	Status                Status
	HighestProposalNumber ProposalNumber
	HighestProposal       Proposal
}

type ReceiveAcceptArgs struct {
}

type ReceiveAcceptReply struct {
}

type ReceiveDecideArgs struct {
}

type ReceiveDecideReply struct {
}
