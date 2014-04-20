package libpaxos

import (
	"distributed2048/gameserver"
	"distributed2048/rpc/paxosrpc"
	"errors"
)

type libpaxos struct {
	allNodes       []PaxosNode
	myPort         int
	decidedHandler func(interface{}) error
}

func NewLibpaxos(allNodes []PaxosNode, port int) (Libpaxos, error) {
	lp := &libpaxos{
		allNodes: allNodes,
		myPort:   port,
	}
	return lp, nil
}

func (lp *libpaxos) ReceiveProposal(args *paxosrpc.ReceiveProposalArgs, reply *paxosrpc.ReceiveProposalReply) error {
	return errors.New("Not implemented yet")
}

func (lp *libpaxos) ReceiveCommit(args *paxosrpc.ReceiveCommitArgs, reply *paxosrpc.ReceiveCommitReply) error {
	return errors.New("Not implemented yet")

}

func (lp *libpaxos) ReceiveDecide(args *paxosrpc.ReceiveDecideArgs, reply *paxosrpc.ReceiveDecideReply) error {
	return errors.New("Not implemented yet")

}

func (lp *libpaxos) Propose(moves []gameserver.Move) error {
	return errors.New("Not implemented yet")

}

func (lp *libpaxos) DecidedHandler(handler func(interface{}) error) {
	lp.decidedHandler = handler
}
