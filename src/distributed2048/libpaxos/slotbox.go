package libpaxos

import (
	"distributed2048/rpc/paxosrpc"
)

type Slot struct {
	Number uint32
	Value  []paxosrpc.Move
}

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

func (sb *SlotBox) Add(slot *Slot) {
	// sb.mutex.Lock()
	sb.slots[slot.Number] = slot
	sb.fastForward()
	// sb.mutex.Unlock()
}

func (sb *SlotBox) Get(number uint32) *Slot {
	// sb.mutex.Lock()
	slot, exists := sb.slots[number]
	// sb.mutex.Unlock()
	if !exists {
		return nil
	}
	return slot
}

func (sb *SlotBox) GetNextUnknownSlotNumber() uint32 {
	// sb.mutex.Lock()
	nextNum := sb.nextUnknownSlotNumber
	// sb.mutex.Unlock()
	return nextNum
}

func (sb *SlotBox) GetNextUnreadSlot() *Slot {
	// sb.mutex.Lock()
	slot, exists := sb.slots[sb.nextUnreadSlotNumber]
	if !exists {
		// sb.mutex.Unlock()
		return nil
	}
	sb.nextUnreadSlotNumber++
	// sb.mutex.Unlock()
	return slot
}

func (sb *SlotBox) fastForward() {
	_, exists := sb.slots[sb.nextUnknownSlotNumber]
	for exists {
		sb.nextUnknownSlotNumber++
		_, exists = sb.slots[sb.nextUnknownSlotNumber]
	}
}
