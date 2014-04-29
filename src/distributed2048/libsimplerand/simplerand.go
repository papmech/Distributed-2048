// Package libsimplerand implements a linear congruential generator that supports manually setting the present x_n value

package libsimplerand

import (
	"sync"
)

const (
	a = uint64(25214903917)
	c = uint64(11)
	m = uint64((1 << 48))
)

type SimpleRand struct {
	xn    uint32
	mutex sync.Mutex
}

func NewSimpleRand(seed uint32) *SimpleRand {
	return &SimpleRand{
		xn: seed,
	}
}

func (r *SimpleRand) Uint32() uint32 {
	r.mutex.Lock()
	r.xn = uint32(((a*uint64(r.xn) + c) % m) >> 16)
	r.mutex.Unlock()
	return r.xn & ((1 << 31) - 1)
}

func (r *SimpleRand) Int() int {
	return int(r.Uint32())
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func (r *SimpleRand) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	return r.Int() % n
}

func (r *SimpleRand) GetCurrent() uint32 {
	r.mutex.Lock()
	ret := r.xn
	r.mutex.Unlock()
	return ret
}

func (r *SimpleRand) SetCurrent(x uint32) {
	r.mutex.Lock()
	r.xn = x
	r.mutex.Unlock()
}
