package libpaxos

import (
	"container/list"
	"distributed2048/rpc/paxosrpc"
	"distributed2048/util"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

const (
	ERROR_LOG bool = true
	DEBUG_LOG bool = true
)

var LOGV = util.NewLogger(DEBUG_LOG, "PAXOS|DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "PAXOS|ERROR", os.Stderr)

type libpaxos struct {
	allNodes       []paxosrpc.Node
	majorityCount  int
	nodeInfo       paxosrpc.Node
	decidedHandler func([]paxosrpc.Move)

	nodesMutex sync.Mutex
	nodes      map[uint32]*node

	// acceptedProposalsMutex sync.Mutex
	// acceptedProposals      []*paxosrpc.Proposal

	globalMutex               sync.Mutex
	highestProposalNumberSeen *paxosrpc.ProposalNumber
	highestAcceptedProposal   *paxosrpc.Proposal
	currentProposal           *paxosrpc.Proposal

	decidedCh  chan *paxosrpc.Proposal
	newValueCh chan []paxosrpc.Move
	// doProposalCh chan *paxosrpc.Proposal

	newValuesQueue     *list.List
	newValuesQueueLock sync.Mutex
}

func NewLibpaxos(nodeID uint32, hostport string, allNodes []paxosrpc.Node) (Libpaxos, error) {
	lp := &libpaxos{
		allNodes:                  allNodes,
		majorityCount:             len(allNodes)/2 + 1,
		nodeInfo:                  paxosrpc.Node{nodeID, hostport},
		nodes:                     make(map[uint32]*node),
		decidedCh:                 make(chan *paxosrpc.Proposal),
		newValueCh:                make(chan []paxosrpc.Move),
		highestProposalNumberSeen: &paxosrpc.ProposalNumber{0, nodeID},
		newValuesQueue:            list.New(),
		// doProposalCh:  make(chan *paxosrpc.Proposal),
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

	go lp.controller()

	return lp, nil
}

func (lp *libpaxos) ReceivePrepare(args *paxosrpc.ReceivePrepareArgs, reply *paxosrpc.ReceivePrepareReply) error {
	// LOGV.Println(lp.nodeInfo.ID, "received prepare", args.ProposalNumber)
	lp.globalMutex.Lock()

	if lp.highestProposalNumberSeen != nil && args.ProposalNumber.LessThan(lp.highestProposalNumberSeen) {
		LOGV.Println(lp.nodeInfo.ID, "rejected prepare", args.ProposalNumber)
		reply.Status = paxosrpc.Reject
	} else {
		LOGV.Println(lp.nodeInfo.ID, "accepted prepare", args.ProposalNumber)
		lp.highestProposalNumberSeen = &args.ProposalNumber
		reply.Status = paxosrpc.OK
		if lp.highestAcceptedProposal != nil {
			reply.HasAcceptedProposal = true
			reply.AcceptedProposal = *lp.highestAcceptedProposal
		} else {
			reply.HasAcceptedProposal = false
		}
	}

	lp.globalMutex.Unlock()
	return nil
}

func (lp *libpaxos) ReceiveAccept(args *paxosrpc.ReceiveAcceptArgs, reply *paxosrpc.ReceiveAcceptReply) error {
	// LOGV.Println(lp.nodeInfo.ID, "received accept", args.Proposal.Number)
	lp.globalMutex.Lock()

	if lp.highestAcceptedProposal != nil && args.Proposal.Number.LessThan(&lp.highestAcceptedProposal.Number) {
		LOGV.Println(lp.nodeInfo.ID, "rejected accept", args.Proposal.Number)
		reply.Status = paxosrpc.Reject
	} else {
		LOGV.Println(lp.nodeInfo.ID, "accepted accept", args.Proposal.Number)
		lp.highestAcceptedProposal = &args.Proposal
		lp.highestProposalNumberSeen = &args.Proposal.Number
		reply.Status = paxosrpc.OK
	}

	lp.globalMutex.Unlock()
	return nil
}

func (lp *libpaxos) ReceiveDecide(args *paxosrpc.ReceiveDecideArgs, reply *paxosrpc.ReceiveDecideReply) error {
	LOGV.Println(lp.nodeInfo.ID, "received decide", args.Proposal.Number)
	lp.globalMutex.Lock()
	lp.highestAcceptedProposal = nil
	lp.globalMutex.Unlock()
	lp.decidedHandler(args.Proposal.Value)
	return nil
}

func (lp *libpaxos) Propose(moves []paxosrpc.Move) error {
	lp.newValueCh <- moves
	return nil
}

func (lp *libpaxos) DecidedHandler(handler func([]paxosrpc.Move)) {
	lp.decidedHandler = handler
}

func (lp *libpaxos) controller() {
	defer LOGV.Println("controller() ending")
	proposalInProgress := false
	doneCh := make(chan struct{})
	for {
		select {
		case moves := <-lp.newValueCh:
			// LOGV.Println("Controller received moves", util.MovesString(moves))
			if proposalInProgress {
				lp.newValuesQueueLock.Lock()
				lp.newValuesQueue.PushBack(moves)
				lp.newValuesQueueLock.Unlock()
			} else {
				proposalInProgress = true
				go lp.doPropose(moves, doneCh)
			}
		case proposal := <-lp.decidedCh:
			lp.decidedHandler(proposal.Value)
		case <-doneCh:
			lp.newValuesQueueLock.Lock()
			if e := lp.newValuesQueue.Front(); e != nil {
				lp.newValuesQueue.Remove(e)
				moves := e.Value.([]paxosrpc.Move)
				go lp.doPropose(moves, doneCh)
			} else {
				proposalInProgress = false
			}
			lp.newValuesQueueLock.Unlock()
		}
	}
}

func (lp *libpaxos) doPropose(moves []paxosrpc.Move, doneCh chan<- struct{}) {
	done := false
	for !done {
		// PHASE 1
		LOGV.Println("Proposer", lp.nodeInfo.ID, ": PHASE 1")
		// LOGV.Println("Proposing", util.MovesString(moves))

		lp.globalMutex.Lock()
		// Make a new proposal such that my_n > n_h
		myProp := paxosrpc.NewProposal(lp.highestProposalNumberSeen.Number+1, lp.nodeInfo.ID, moves)
		lp.highestProposalNumberSeen = &myProp.Number
		// lp.currentProposal = myProp // TODO: Need this?
		lp.globalMutex.Unlock()

		// Send proposal to everybody
		promisedCount := 1
		var otherProposal *paxosrpc.Proposal
		lp.nodesMutex.Lock()
		for _, node := range lp.nodes {
			if node.Info.ID == lp.nodeInfo.ID {
				continue // skip myself
			}

			client := node.getRPCClient()
			if client == nil {
				continue
			}

			args := &paxosrpc.ReceivePrepareArgs{lp.nodeInfo, myProp.Number}
			var reply paxosrpc.ReceivePrepareReply
			err := client.Call("PaxosNode.ReceivePrepare", args, &reply)
			if err != nil {
				node.Client = nil // so it will try to redial in future attempts
				continue          // skip nodes that are unreachable
			}

			lp.globalMutex.Lock()
			if reply.Status == paxosrpc.OK {
				if reply.HasAcceptedProposal {
					if otherProposal == nil || reply.AcceptedProposal.Number.GreaterThan(&otherProposal.Number) {
						otherProposal = &reply.AcceptedProposal
					}
				}
				promisedCount++
			} else {
				// do nothing if REJECTED
			}
			lp.globalMutex.Unlock()
		}
		lp.nodesMutex.Unlock()

		// Got majority?
		if promisedCount < lp.majorityCount {
			LOGV.Println(lp.nodeInfo.ID, " couldn't get a majority. Got", promisedCount, "needed", lp.majorityCount)
			// Backoff
			num := time.Duration(rand.Int()%100 + 25)
			time.Sleep(num * time.Millisecond)
			continue // try again
		}

		propToAccept := otherProposal
		if otherProposal == nil {
			propToAccept = myProp
		}
		LOGV.Println(lp.nodeInfo.ID, "got a majority on", propToAccept.Number.String())

		// PHASE 2
		LOGV.Println("Proposer", lp.nodeInfo.ID, ": PHASE 2")

		// Send <accept, myn, V> to all nodes
		acceptedCount := 1
		lp.nodesMutex.Lock()
		for _, node := range lp.nodes {
			if node.Info.ID == lp.nodeInfo.ID {
				continue // skip myself
			}

			client := node.getRPCClient()
			if client == nil {
				continue
			}

			args := &paxosrpc.ReceiveAcceptArgs{*propToAccept}
			var reply paxosrpc.ReceiveAcceptReply
			err := client.Call("PaxosNode.ReceiveAccept", args, &reply)
			if err != nil {
				continue // skip nodes that are unreachable
			}

			if reply.Status == paxosrpc.OK {
				acceptedCount++
			} else {
				// do nothing if REJECTED
			}
		}
		lp.nodesMutex.Unlock()

		// Got majority?
		if acceptedCount < lp.majorityCount {
			// TODO: do I nil-out the lp.highestAcceptedProposal etc?
			// TODO: add backoff
			continue // try again
		}

		// Send <decide, va> to all nodes
		lp.nodesMutex.Lock()
		for _, node := range lp.nodes {
			if node.Info.ID == lp.nodeInfo.ID {
				continue // skip myself
			}

			client := node.getRPCClient()
			if client == nil {
				continue
			}

			args := &paxosrpc.ReceiveDecideArgs{*propToAccept}
			var reply paxosrpc.ReceiveDecideReply
			err := client.Call("PaxosNode.ReceiveDecide", args, &reply)
			if err != nil {
				continue // skip nodes that are unreachable
			}
		}
		lp.nodesMutex.Unlock()

		lp.globalMutex.Lock()
		lp.highestAcceptedProposal = nil
		lp.globalMutex.Unlock()
		lp.decidedHandler(propToAccept.Value)

		if propToAccept.Number.Equal(&myProp.Number) {
			done = true
		}
	}

	doneCh <- struct{}{}
}
