package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

type libpaxos struct {
	allNodes       []paxosrpc.Node
	majorityCount  int
	nodeInfo       paxosrpc.Node
	decidedHandler func([]paxosrpc.Move) error

	nodesMutex sync.Mutex
	nodes      map[uint32]*node

	acceptedProposalsMutex sync.Mutex
	acceptedProposals      []paxosrpc.Proposal

	globalMutex               sync.Mutex
	highestProposalNumberSeen paxosrpc.ProposalNumber
	highestProposal           paxosrpc.Proposal
	currentProposal           *paxosrpc.Proposal

	doneCh chan struct{}
}

func NewLibpaxos(nodeID uint32, hostport string, allNodes []paxosrpc.Node) (Libpaxos, error) {
	lp := &libpaxos{
		allNodes:      allNodes,
		majorityCount: len(allNodes)/2 + 1,
		nodeInfo:      paxosrpc.Node{nodeID, hostport},
		nodes:         make(map[uint32]*node),
	}

	for _, node := range allNodes {
		lp.nodes[node.ID] = NewNode(node)
	}

	// Start the RPC handlers
	rpc.RegisterName("PaxosNode", paxosrpc.Wrap(lp))
	rpc.HandleHTTP()
	l, err := net.Listen("tcp", fmt.Sprintf(hostport))
	if err != nil {
		return nil, err
	}
	go http.Serve(l, nil)

	return lp, nil
}

func (lp *libpaxos) ReceivePrepare(args *paxosrpc.ReceivePrepareArgs, reply *paxosrpc.ReceivePrepareReply) error {
	lp.globalMutex.Lock()

	if args.Proposal.Number.LessThan(lp.highestProposalNumberSeen) {
		reply.Status = paxosrpc.Reject
		reply.HighestProposalNumber = lp.highestProposalNumberSeen
	} else {
		reply.Status = paxosrpc.Accept
		if args.Proposal.Number.GreaterThan(lp.highestProposalNumberSeen) {
			lp.highestProposalNumberSeen = args.Proposal.Number
			lp.highestProposal = args.Proposal
		}
		reply.HighestProposal = lp.highestProposal
	}

	lp.globalMutex.Unlock()
	return errors.New("Not implemented yet")
}

func (lp *libpaxos) ReceiveAccept(args *paxosrpc.ReceiveAcceptArgs, reply *paxosrpc.ReceiveAcceptReply) error {
	return errors.New("Not implemented yet")
}

func (lp *libpaxos) ReceiveDecide(args *paxosrpc.ReceiveDecideArgs, reply *paxosrpc.ReceiveDecideReply) error {
	return errors.New("Not implemented yet")
}

func (lp *libpaxos) Propose(moves []paxosrpc.Move) error {
	done := false
	for !done {
		// PHASE 1

		lp.globalMutex.Lock()
		// Make a new proposal such that my_n > n_h
		prop := paxosrpc.NewProposal(lp.highestProposalNumberSeen.Number+1, lp.nodeInfo.ID, moves)
		lp.currentProposal = prop
		lp.globalMutex.Unlock()

		// Send proposal to everybody
		acceptedCount := 0
		lp.nodesMutex.Lock()
		for _, node := range lp.nodes {
			client := node.getRPCClient()
			if client == nil {
				continue
			}

			args := &paxosrpc.ReceivePrepareArgs{lp.nodeInfo, *prop}
			var reply paxosrpc.ReceivePrepareReply
			err := client.Call("PaxosNode.ReceivePrepare", args, &reply)
			if err != nil {
				continue // skip nodes that are unreachable
			}

			if reply.Status == paxosrpc.Reject {
				lp.globalMutex.Lock()
				lp.highestProposalNumberSeen = reply.HighestProposalNumber
				lp.globalMutex.Unlock()
			} else {
				acceptedCount++
			}
		}
		lp.nodesMutex.Unlock()

		if acceptedCount < lp.majorityCount {
			continue // try again
		}

	}

	// Block until we send out all the Decides
	<-lp.doneCh

	return nil
}

// func (lp *libpaxos) DecidedHandler(handler func([]paxosrpc.Move) error) {
// 	lp.decidedHandler = handler
// }
