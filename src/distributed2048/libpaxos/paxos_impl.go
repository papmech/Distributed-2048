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
	ERROR_LOG          bool = false
	DEBUG_LOG          bool = false
	DUMP_SLOTS         bool = false
	SHOW_DECIDED_SLOTS bool = false
)

var LOGV = util.NewLogger(DEBUG_LOG, "PAXOS|DEBUG", os.Stdout)
var LOGE = util.NewLogger(ERROR_LOG, "PAXOS|ERROR", os.Stderr)

type libpaxos struct {
	allNodes       []paxosrpc.Node
	majorityCount  int // minimum number of nodes for a majority to be reached
	myNode         paxosrpc.Node
	decidedHandler func(*paxosrpc.ProposalValue) // called when decided value received after successful paxos round

	nodesMutex sync.Mutex
	nodes      map[uint32]*node

	dataMutex                 sync.Mutex
	highestProposalNumberSeen *paxosrpc.ProposalNumber
	highestAcceptedProposal   *paxosrpc.Proposal

	slotBox      *SlotBox   // holds previously decided slots
	slotBoxMutex sync.Mutex // lock for slotBox

	triggerHandlerCallCh chan struct{}
	newValueCh           chan *paxosrpc.ProposalValue

	newValuesQueue     *list.List // Queue for new values to be later proposed
	newValuesQueueLock sync.Mutex // Queue lock

	interruptFunc func(id uint32, action PaxosAction, slotNumber uint32)
}

func NewLibpaxos(nodeID uint32, hostport string, allNodes []paxosrpc.Node) (Libpaxos, error) {
	lp := &libpaxos{
		allNodes:                  allNodes,
		majorityCount:             len(allNodes)/2 + 1,
		myNode:                    paxosrpc.Node{nodeID, hostport},
		nodes:                     make(map[uint32]*node),
		newValueCh:                make(chan *paxosrpc.ProposalValue),
		highestProposalNumberSeen: &paxosrpc.ProposalNumber{0, nodeID},
		slotBox:                   NewSlotBox(),
		triggerHandlerCallCh:      make(chan struct{}, 1000),
		newValuesQueue:            list.New(),
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
	if DUMP_SLOTS { // Writes the contents of slotbox to file at fixed intervals
		go lp.dumpSlots()
	}

	return lp, nil
}

func (lp *libpaxos) SetInterruptFunc(f func(id uint32, action PaxosAction, slotNumber uint32)) {
	lp.interruptFunc = f
}

func (lp *libpaxos) ReceivePrepare(args *paxosrpc.ReceivePrepareArgs, reply *paxosrpc.ReceivePrepareReply) error {
	if lp.interruptFunc != nil {
		lp.interruptFunc(lp.myNode.ID, Prepare, lp.slotBox.nextUnknownSlotNumber)
	}

	lp.dataMutex.Lock()

	lp.slotBoxMutex.Lock()
	slot := lp.slotBox.Get(args.CommandSlotNumber)
	lp.slotBoxMutex.Unlock()
	if slot != nil {
		// The Proposer will suggest a slot number for its proposal. If that slot
		// number has already been decided upon, tell the Proposer, and give it
		// the decided value, so it can update its own slot box and choose a
		// different slot number.
		reply.Status = paxosrpc.DecidedValueExists
		reply.DecidedSlotNumber = slot.Number
		reply.DecidedValue = *slot.Value
	} else if lp.highestProposalNumberSeen != nil && args.ProposalNumber.LessThan(lp.highestProposalNumberSeen) {
		// If the proposal number is not highest, reject automatically.
		reply.Status = paxosrpc.Reject
	} else {
		// If the proposal number is highest, then OK it, but also send back
		// any accepted proposals.
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
	if lp.interruptFunc != nil {
		lp.interruptFunc(lp.myNode.ID, Accept, lp.slotBox.nextUnknownSlotNumber)
	}
	lp.dataMutex.Lock()

	lp.slotBoxMutex.Lock()
	slot := lp.slotBox.Get(args.Proposal.CommandSlotNumber)
	lp.slotBoxMutex.Unlock()

	// Only accept a proposal if it has the highest proposal number so far.
	if args.Proposal.Number.LessThan(lp.highestProposalNumberSeen) {
		reply.Status = paxosrpc.Reject
	} else if lp.highestAcceptedProposal != nil && (args.Proposal.Number.LessThan(&lp.highestAcceptedProposal.Number)) {
		reply.Status = paxosrpc.Reject
	} else if slot != nil {
		// Someone else already filled this slot!
		reply.Status = paxosrpc.Reject
	} else {
		lp.highestAcceptedProposal = &args.Proposal
		lp.highestProposalNumberSeen = &args.Proposal.Number
		reply.Status = paxosrpc.OK
	}

	lp.dataMutex.Unlock()
	return nil
}

func (lp *libpaxos) ReceiveDecide(args *paxosrpc.ReceiveDecideArgs, reply *paxosrpc.ReceiveDecideReply) error {
	if lp.interruptFunc != nil {
		lp.interruptFunc(lp.myNode.ID, Decide, lp.slotBox.nextUnknownSlotNumber)
	}
	lp.dataMutex.Lock()

	// Send the proposal to the slot box.
	lp.slotBoxMutex.Lock()
	lp.slotBox.Add(NewSlot(args.Proposal.CommandSlotNumber, &args.Proposal.Value))
	lp.slotBoxMutex.Unlock()

	// Trigger the go routine which will check if any values can be sent to
	// the handler.
	lp.triggerHandlerCallCh <- struct{}{}

	// Reset the paxos state for the next round of paxos
	lp.highestAcceptedProposal = nil

	lp.dataMutex.Unlock()
	return nil
}

func (lp *libpaxos) Propose(proposal *paxosrpc.ProposalValue) error {
	lp.newValueCh <- proposal
	return nil
}

func (lp *libpaxos) DecidedHandler(handler func(proposal *paxosrpc.ProposalValue)) {
	lp.decidedHandler = handler
}

// handleCaller is triggered by sending to triggerHandlerCallCh to check if
// the next unread slot can be sent to the handler.
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
					if SHOW_DECIDED_SLOTS {
						fmt.Println("Node", lp.myNode.ID, "got slot", slot.Number)
					}
					lp.decidedHandler(slot.Value)
				}
			}
		}
	}
}

// controller handles the arrival of new proposal values, and also what
// happens when a paxos round is complete.
//
// New values should be queue if a paxos round is in progress.
//
// When a paxos round completes, the next value in the queue is proposed.
func (lp *libpaxos) controller() {
	proposalInProgress := false
	doneCh := make(chan struct{})
	for {
		select {
		case proposal := <-lp.newValueCh:
			if proposalInProgress {
				LOGV.Println("Proposal in progress, deferring")
				lp.newValuesQueueLock.Lock()
				lp.newValuesQueue.PushBack(proposal)
				lp.newValuesQueueLock.Unlock()
			} else {
				proposalInProgress = true
				go lp.doPropose(proposal, doneCh)
			}
		case <-doneCh:
			lp.newValuesQueueLock.Lock()
			if e := lp.newValuesQueue.Front(); e != nil {
				lp.newValuesQueue.Remove(e)
				proposal := e.Value.(*paxosrpc.ProposalValue)
				go lp.doPropose(proposal, doneCh)
			} else {
				proposalInProgress = false
			}
			lp.newValuesQueueLock.Unlock()
		}
	}
}

// doPropose will keep attempting to submit a proposal to the Paxos cluster
// with the given value. This may not always succeed because there will be
// competing proposals, but when doPropose returns, it guarantees that the
// value has been decided in a quorum.
func (lp *libpaxos) doPropose(value *paxosrpc.ProposalValue, doneCh chan<- struct{}) {
	LOGV.Println("doing propose")
	done := false // true if $moves has been Decided, false otherwise.
	for !done {
		retry := false // true if we should try again, for whatever reason, false otherwise.

		// PHASE 1
		LOGV.Println("Proposer", lp.myNode.ID, ": PHASE 1")

		// Make a new proposal such that my_n > n_h
		lp.dataMutex.Lock()
		lp.slotBoxMutex.Lock()
		myProp := paxosrpc.NewProposal(lp.highestProposalNumberSeen.Number+1, lp.slotBox.GetNextUnknownSlotNumber(), lp.myNode.ID, *value)
		lp.highestProposalNumberSeen = &myProp.Number
		lp.slotBoxMutex.Unlock()
		lp.dataMutex.Unlock()

		// Send proposal to everybody
		promisedCount := 1 // include myself
		var otherProposal *paxosrpc.Proposal
		lp.nodesMutex.Lock()
		for _, node := range lp.nodes {
			if node.Info.ID == lp.myNode.ID {
				continue // skip myself
			}

			client := node.getRPCClient()
			if client == nil {
				continue
			}

			args := &paxosrpc.ReceivePrepareArgs{lp.myNode, myProp.Number, myProp.CommandSlotNumber}
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
				lp.slotBox.Add(NewSlot(reply.DecidedSlotNumber, &reply.DecidedValue))
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
			LOGV.Println(lp.myNode.ID, " couldn't get a majority. Got", promisedCount, "needed", lp.majorityCount)
			// Backoff
			num := time.Duration(rand.Int()%100 + 25)
			time.Sleep(num * time.Millisecond)
			continue // try again
		}

		propToAccept := otherProposal
		if otherProposal == nil {
			propToAccept = myProp
		}
		LOGV.Println(lp.myNode.ID, "got a majority on", propToAccept.Number.String())

		// PHASE 2
		LOGV.Println("Proposer", lp.myNode.ID, ": PHASE 2")

		// Send <accept, myn, V> to all nodes
		acceptedCount := 1
		lp.nodesMutex.Lock()
		for _, node := range lp.nodes {
			if node.Info.ID == lp.myNode.ID {
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
				LOGE.Println(err)
				node.Client = nil // so it will try to redial in future attempts
				continue          // skip nodes that are unreachable
			} else if timedOut {
				LOGE.Println("RPC call PaxosNode.ReceiveAccept to", node.Info.ID, "timed out")
			}

			if reply.Status == paxosrpc.OK {
				acceptedCount++
			} else {
				LOGV.Println("Node", node.Info.ID, "rejected my Accept message with status", reply.Status)
			} // do nothing if REJECTED
		}
		lp.nodesMutex.Unlock()

		// Got majority?
		if acceptedCount < lp.majorityCount {
			LOGV.Println("Couldn't get majority for PHASE 2,", acceptedCount, "/", lp.majorityCount)
			// Backoff
			num := time.Duration(rand.Int()%100 + 25)
			time.Sleep(num * time.Millisecond)
			continue // try again
		}

		LOGV.Printf("%d got majority (%d/%d) on ACCEPT for slot %d, seqnum is %s, value is\n%s\n", lp.myNode.ID, acceptedCount, lp.majorityCount, propToAccept.CommandSlotNumber, propToAccept.Number.String(), util.MovesString(propToAccept.Value.Moves))

		// Send <decide, va> to all nodes
		lp.nodesMutex.Lock()
		for _, node := range lp.nodes {
			if node.Info.ID == lp.myNode.ID {
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
		}

		LOGV.Printf("%d decided on slot %d, seqnum is %s, value is\n%s\n", lp.myNode.ID, propToAccept.CommandSlotNumber, propToAccept.Number.String(), util.MovesString(propToAccept.Value.Moves))
		lp.nodesMutex.Unlock()

		lp.dataMutex.Lock()
		lp.highestAcceptedProposal = nil
		lp.dataMutex.Unlock()

		lp.slotBoxMutex.Lock()
		lp.slotBox.Add(NewSlot(propToAccept.CommandSlotNumber, &propToAccept.Value))
		lp.slotBoxMutex.Unlock()

		lp.triggerHandlerCallCh <- struct{}{}

		if propToAccept.Number.Equal(&myProp.Number) {
			done = true
		}
	}
	LOGV.Println("Done propose")
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

// dumpSlots writes the
func (lp *libpaxos) dumpSlots() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:

			file, _ := os.Create(fmt.Sprintf("%d_slots_%s.log", lp.myNode.ID, time.Now().String()))
			lp.slotBoxMutex.Lock()
			file.WriteString(lp.slotBox.String())
			lp.slotBoxMutex.Unlock()
			file.Close()
		}
	}
}
