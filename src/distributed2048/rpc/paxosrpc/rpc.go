package paxosrpc

type RemotePaxosNode interface {
	ReceivePrepare(*ReceivePrepareArgs, *ReceivePrepareReply) error
	ReceiveAccept(*ReceiveAcceptArgs, *ReceiveAcceptReply) error
	ReceiveDecide(*ReceiveDecideArgs, *ReceiveDecideReply) error
}

type PaxosNode struct {
	RemotePaxosNode
}

func Wrap(pn RemotePaxosNode) RemotePaxosNode {
	return &PaxosNode{pn}
}
