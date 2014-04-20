package paxosrpc

type Status int

const (
	OK Status = iota + 1
)

type ReceiveProposalArgs struct {
}

type ReceiveProposalReply struct {
	Status Status
}

type ReceiveCommitArgs struct {
}

type ReceiveCommitReply struct {
}

type ReceiveDecideArgs struct {
}

type ReceiveDecideReply struct {
}
