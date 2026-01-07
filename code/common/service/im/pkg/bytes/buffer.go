package bytes

import "sync"

type Buffer struct {
	buf  []byte
	next *Buffer
}

func (b *Buffer) Bytes() []byte {
	return b.buf
}

type Pool struct {
	lock sync.Mutex
	free *Buffer
	max  int
	num  int
	size int
}

func NewPool(num, size int) (p *Pool) {
	p = new(Pool)
	p.init(num, size)
	return
}

// Init init the memory buffer.
func (p *Pool) Init(num, size int) {
	p.init(num, size)
}

func (p *Pool) init(num, size int) {
	p.num = num
	p.size = size
	p.max = num * size
	p.grow()
}

func (p *Pool) grow() {
	var (
		i   int
		cur *Buffer
		buf []byte
		bs  []Buffer
	)
	buf = make([]byte, p.max)
	bs = make([]Buffer, p.num)
	p.free = &bs[0]
	cur = p.free
	for i = 1; i < p.num; i++ {
		cur.buf = buf[(i-1)*p.size : i*p.size]
		cur.next = &bs[i]
		cur = cur.next
	}
	cur.buf = buf[(i-1)*p.size : i*p.size]
	cur.next = nil
}

func (p *Pool) Get() (b *Buffer) {
	p.lock.Lock()
	if b = p.free; b == nil {
		p.grow()
		b = p.free
	}
	p.free = b.next
	p.lock.Unlock()
	return
}

func (p *Pool) Put(b *Buffer) {
	p.lock.Lock()
	b.next = p.free
	p.free = b
	p.lock.Unlock()
}
