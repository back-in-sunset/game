package session

import "fmt"

var (
	ErrRingEmpty = fmt.Errorf("ring empty")
	ErrRingFull  = fmt.Errorf("ring full")
)

type Ring struct {
	rp   uint64
	wp   uint64
	num  uint64
	mask uint64
	data [][]byte
}

func NewRing(size int) *Ring {
	r := &Ring{}
	r.Init(size)
	return r
}

func (r *Ring) Init(size int) {
	if size <= 0 {
		size = 1
	}
	n := uint64(size)
	if n&(n-1) != 0 {
		n = nextPowerOfTwo(n)
	}
	r.data = make([][]byte, n)
	r.num = n
	r.mask = n - 1
	r.rp = 0
	r.wp = 0
}

func (r *Ring) Len() int {
	return int(r.wp - r.rp)
}

func (r *Ring) Cap() int {
	return int(r.num)
}

func (r *Ring) Push(payload []byte) error {
	if r.wp-r.rp >= r.num {
		return ErrRingFull
	}
	idx := r.wp & r.mask
	r.data[idx] = payload
	r.wp++
	return nil
}

func (r *Ring) ForcePush(payload []byte) (dropped []byte) {
	if r.wp-r.rp >= r.num {
		idx := r.rp & r.mask
		dropped = r.data[idx]
		r.data[idx] = nil
		r.rp++
	}
	_ = r.Push(payload)
	return dropped
}

func (r *Ring) Pop() ([]byte, error) {
	if r.rp == r.wp {
		return nil, ErrRingEmpty
	}
	idx := r.rp & r.mask
	payload := r.data[idx]
	r.data[idx] = nil
	r.rp++
	return payload, nil
}

func nextPowerOfTwo(v uint64) uint64 {
	v--
	v |= v >> 1
	v |= v >> 2
	v |= v >> 4
	v |= v >> 8
	v |= v >> 16
	v |= v >> 32
	v++
	return v
}
