package paxosrpc

type Status int

const (
	OK Status = iota + 1
)

type ReceivePrepareArgs struct {
	NodeID   int
	Proposal Proposal
}

type ReceivePrepareReply struct {
	Status Status
}

type ReceiveAcceptArgs struct {
}

type ReceiveAcceptReply struct {
}

type ReceiveDecideArgs struct {
}

type ReceiveDecideReply struct {
}
