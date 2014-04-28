package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
)

type Slot struct {
	Number uint32
	Value  []paxosrpc.Move
}

// SlotBox stores the history of all decided proposals.
type SlotBox struct {
	slots                 map[uint32]*Slot
	nextUnreadSlotNumber  uint32
	nextUnknownSlotNumber uint32
}

func NewSlotBox() *SlotBox {
	return &SlotBox{make(map[uint32]*Slot), 0, 0}
}

func NewSlot(number uint32, value []paxosrpc.Move) *Slot {
	return &Slot{number, value}
}

// Add puts the given slot into the slotbox, and fastforwards the next unknown
// slot number forward if necessary.
func (sb *SlotBox) Add(slot *Slot) {
	sb.slots[slot.Number] = slot
	sb.fastForward()
}

// Gets a specific slot, and returns nil if it does not exist
func (sb *SlotBox) Get(number uint32) *Slot {
	slot, exists := sb.slots[number]
	if !exists {
		return nil
	}
	return slot
}

// Gets the number of next unfilled slot
func (sb *SlotBox) GetNextUnknownSlotNumber() uint32 {
	nextNum := sb.nextUnknownSlotNumber
	return nextNum
}

// Gets the slot that has not yet been read
func (sb *SlotBox) GetNextUnreadSlot() *Slot {
	slot, exists := sb.slots[sb.nextUnreadSlotNumber]
	if !exists {
		return nil
	}
	sb.nextUnreadSlotNumber++
	return slot
}

func (sb *SlotBox) fastForward() {
	_, exists := sb.slots[sb.nextUnknownSlotNumber]
	for exists {
		sb.nextUnknownSlotNumber++
		_, exists = sb.slots[sb.nextUnknownSlotNumber]
	}
}
