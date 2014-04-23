package paxosrpc

type RemotePaxosNode interface {
	// ReceiveProposal is called by a Proposer via RPC when it wishes to
	// propose a new value for Paxos. This first checks if the desired slot
	// has already been previous Decided upon, and if so sends a rejection
	// (DecidedValueExists) with the decided slot number and value. If not,
	// checks that the incoming proposal number is higher than any we've seen,
	// and sends an OK, with the highest Accepted proposal thus far.
	// Otherwise, sends a rejection (Reject).
	ReceivePrepare(args *ReceivePrepareArgs, reply *ReceivePrepareReply) error
	// ReceiveAccept is called by a Proposer via RPC when it wishes to ask all
	// other nodes if they accept the proposed value. This checks that the
	// incoming proposal number is equal to or higher than the highest
	// Accepted proposal number. If so, sends OK, otherwise Reject.
	ReceiveAccept(args *ReceiveAcceptArgs, reply *ReceiveAcceptReply) error
	// ReceiveDecide is called by a Proposer via RPC when one round of Paxos
	// has completed, and everyone has agreed on a value. This will
	// asynchronously trigger the onDecided handler if it exists.
	ReceiveDecide(args *ReceiveDecideArgs, reply *ReceiveDecideReply) error
}

type PaxosNode struct {
	RemotePaxosNode
}

func Wrap(pn RemotePaxosNode) RemotePaxosNode {
	return &PaxosNode{pn}
}
