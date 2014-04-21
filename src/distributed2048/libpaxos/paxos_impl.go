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
	RPC_TIMEOUT_MILLISEC = 500
)

const (
	ERROR_LOG bool = false
	DEBUG_LOG bool = false
)

var LOGV = util.NewLogger(DEBUG_LOG, "PAXOS|DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "PAXOS|ERROR", os.Stderr)

type libpaxos struct {
	allNodes       []paxosrpc.Node
	majorityCount  int
	nodeInfo       paxosrpc.Node
	decidedHandler func(uint32, []paxosrpc.Move)

	nodesMutex sync.Mutex
	nodes      map[uint32]*node

	// acceptedProposalsMutex sync.Mutex
	// acceptedProposals      []*paxosrpc.Proposal

	dataMutex                 sync.Mutex
	highestProposalNumberSeen *paxosrpc.ProposalNumber
	highestAcceptedProposal   *paxosrpc.Proposal
	slotBox                   *SlotBox
	slotBoxMutex              sync.Mutex
	triggerHandlerCallCh      chan struct{}
	// currentProposal           *paxosrpc.Proposal
	// slotNumbers           map[uint32]bool
	// nextCommandSlotNumber uint32

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
		slotBox:                   NewSlotBox(),
		triggerHandlerCallCh:      make(chan struct{}, 1000),
		// slotNumbers:               make(map[uint32]bool),
		// nextCommandSlotNumber:     0,
		newValuesQueue: list.New(),
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
	go lp.handleCaller()

	return lp, nil
}

func (lp *libpaxos) ReceivePrepare(args *paxosrpc.ReceivePrepareArgs, reply *paxosrpc.ReceivePrepareReply) error {
	// LOGV.Println(lp.nodeInfo.ID, "received prepare", args.ProposalNumber)
	lp.dataMutex.Lock()

	// TODO: Remove this
	// Randomly simulate time out
	doTimeout := rand.Int()%100 < 3 // 20% of the time
	if doTimeout {
		LOGV.Println(lp.nodeInfo.ID, "simulating 15 second interruption.")
		time.Sleep(15 * time.Second)
	}

	lp.slotBoxMutex.Lock()
	slot := lp.slotBox.Get(args.CommandSlotNumber)
	lp.slotBoxMutex.Unlock()
	if slot != nil {
		reply.Status = paxosrpc.DecidedValueExists
		reply.DecidedSlotNumber = slot.Number
		reply.DecidedValue = slot.Value
	} else if lp.highestProposalNumberSeen != nil && args.ProposalNumber.LessThan(lp.highestProposalNumberSeen) {
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

	lp.dataMutex.Unlock()
	return nil
}

func (lp *libpaxos) ReceiveAccept(args *paxosrpc.ReceiveAcceptArgs, reply *paxosrpc.ReceiveAcceptReply) error {
	// LOGV.Println(lp.nodeInfo.ID, "received accept", args.Proposal.Number)
	lp.dataMutex.Lock()

	if lp.highestAcceptedProposal != nil && args.Proposal.Number.LessThan(&lp.highestAcceptedProposal.Number) {
		LOGV.Println(lp.nodeInfo.ID, "rejected accept", args.Proposal.Number)
		reply.Status = paxosrpc.Reject
	} else {
		LOGV.Println(lp.nodeInfo.ID, "accepted accept", args.Proposal.Number)
		lp.highestAcceptedProposal = &args.Proposal
		lp.highestProposalNumberSeen = &args.Proposal.Number
		reply.Status = paxosrpc.OK
	}

	lp.dataMutex.Unlock()
	return nil
}

func (lp *libpaxos) ReceiveDecide(args *paxosrpc.ReceiveDecideArgs, reply *paxosrpc.ReceiveDecideReply) error {
	LOGV.Println(lp.nodeInfo.ID, "received decide", args.Proposal.Number, "for slot", args.Proposal.CommandSlotNumber)
	lp.dataMutex.Lock()
	lp.highestAcceptedProposal = nil
	lp.slotBoxMutex.Lock()
	lp.slotBox.Add(NewSlot(args.Proposal.CommandSlotNumber, args.Proposal.Value))
	lp.slotBoxMutex.Unlock()
	lp.triggerHandlerCallCh <- struct{}{}
	lp.dataMutex.Unlock()
	// TODO: only call the handler IF we obtain the next slot number in line
	// lp.decidedHandler(args.Proposal.CommandSlotNumber, args.Proposal.Value)
	return nil
}

func (lp *libpaxos) Propose(moves []paxosrpc.Move) error {
	lp.newValueCh <- moves
	return nil
}

func (lp *libpaxos) DecidedHandler(handler func(slotNumber uint32, moves []paxosrpc.Move)) {
	lp.decidedHandler = handler
}

func (lp *libpaxos) handleCaller() {
	for {
		select {
		case <-lp.triggerHandlerCallCh:
			for {
				lp.slotBoxMutex.Lock()
				slot := lp.slotBox.GetNextUnreadSlot()
				lp.slotBoxMutex.Unlock()
				if slot == nil {
					break
				}
				if lp.decidedHandler != nil {
					lp.decidedHandler(slot.Number, slot.Value)
				}
			}
		}
	}
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
		// case proposal := <-lp.decidedCh:
		// lp.decidedHandler(proposal.CommandSlotNumber, proposal.Value)
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
		// TODO: Remove this
		// Articifically stagger the proposers
		// time.Sleep(time.Duration(rand.Int()%1000) * time.Millisecond)

		retry := false

		// PHASE 1
		LOGV.Println("Proposer", lp.nodeInfo.ID, ": PHASE 1")
		// LOGV.Println("Proposing", util.MovesString(moves))

		lp.dataMutex.Lock()
		// Make a new proposal such that my_n > n_h
		lp.slotBoxMutex.Lock()
		myProp := paxosrpc.NewProposal(lp.highestProposalNumberSeen.Number+1, lp.slotBox.GetNextUnknownSlotNumber(), lp.nodeInfo.ID, moves)
		lp.highestProposalNumberSeen = &myProp.Number
		// lp.currentProposal = myProp // TODO: Need this?
		lp.slotBoxMutex.Unlock()
		lp.dataMutex.Unlock()

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

			args := &paxosrpc.ReceivePrepareArgs{lp.nodeInfo, myProp.Number, myProp.CommandSlotNumber}
			var reply paxosrpc.ReceivePrepareReply
			reply.Status = paxosrpc.Reject

			timedOut, err := rpcCallWithTimeout(client, "PaxosNode.ReceivePrepare", args, &reply)
			if err != nil {
				node.Client = nil // so it will try to redial in future attempts
				continue          // skip nodes that are unreachable
			} else if timedOut {
				LOGE.Println("RPC call PaxosNode.ReceivePrepare to", node.Info.ID, "timed out")
			}

			lp.dataMutex.Lock()
			switch reply.Status {
			case paxosrpc.OK:
				if reply.HasAcceptedProposal {
					if otherProposal == nil || reply.AcceptedProposal.Number.GreaterThan(&otherProposal.Number) {
						otherProposal = &reply.AcceptedProposal
					}
				}
				promisedCount++

			case paxosrpc.DecidedValueExists:
				// Oops, better fill in that value
				lp.slotBoxMutex.Lock()
				lp.slotBox.Add(NewSlot(reply.DecidedSlotNumber, reply.DecidedValue))
				lp.slotBoxMutex.Unlock()

				lp.triggerHandlerCallCh <- struct{}{}

				retry = true
				break

			case paxosrpc.Reject:
				// do nothing if REJECTED
			}
			lp.dataMutex.Unlock()
		}
		lp.nodesMutex.Unlock()

		// Retry?
		if retry {
			continue
		}

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
			reply.Status = paxosrpc.Reject

			timedOut, err := rpcCallWithTimeout(client, "PaxosNode.ReceiveAccept", args, &reply)
			if err != nil {
				node.Client = nil // so it will try to redial in future attempts
				continue          // skip nodes that are unreachable
			} else if timedOut {
				LOGE.Println("RPC call PaxosNode.ReceiveAccept to", node.Info.ID, "timed out")
			}

			if reply.Status == paxosrpc.OK {
				acceptedCount++
			} // do nothing if REJECTED
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

			rpcCallWithTimeout(client, "PaxosNode.ReceiveDecide", args, &reply)
			// We don't care if that rpc call had an error or timed out

			// err := client.Call("PaxosNode.ReceiveDecide", args, &reply)
			// if err != nil {
			// 	continue // skip nodes that are unreachable
			// }
		}
		lp.nodesMutex.Unlock()

		lp.dataMutex.Lock()
		lp.highestAcceptedProposal = nil
		lp.dataMutex.Unlock()

		lp.slotBoxMutex.Lock()
		lp.slotBox.Add(NewSlot(propToAccept.CommandSlotNumber, propToAccept.Value))
		lp.slotBoxMutex.Unlock()
		lp.triggerHandlerCallCh <- struct{}{}

		if propToAccept.Number.Equal(&myProp.Number) {
			done = true
		}
	}

	doneCh <- struct{}{}
}

func rpcCallWithTimeout(client *rpc.Client, serviceMethod string, args, reply interface{}) (bool, error) {
	ch := make(chan error, 1)
	go func() { ch <- client.Call(serviceMethod, args, reply) }()
	select {
	case err := <-ch:
		return false, err
	case <-time.After(RPC_TIMEOUT_MILLISEC * time.Millisecond):
		return true, nil
	}
	return false, nil
}
