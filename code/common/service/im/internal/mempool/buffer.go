package mempool

import (
	"bytes"
	"sync"
)

type BytePool struct {
	size int
	pool sync.Pool
}

func NewBytePool(size int) *BytePool {
	p := &BytePool{size: size}
	p.pool.New = func() any {
		buf := make([]byte, size)
		return &buf
	}
	return p
}

func (p *BytePool) Get() []byte {
	if p == nil {
		return nil
	}
	bufPtr := p.pool.Get().(*[]byte)
	buf := *bufPtr
	return buf[:cap(buf)]
}

func (p *BytePool) Put(buf []byte) {
	if p == nil || buf == nil {
		return
	}
	if cap(buf) < p.size {
		return
	}
	buf = buf[:p.size]
	p.pool.Put(&buf)
}

type BufferPool struct {
	size int
	pool sync.Pool
}

func NewBufferPool(size int) *BufferPool {
	p := &BufferPool{size: size}
	p.pool.New = func() any {
		return bytes.NewBuffer(make([]byte, 0, size))
	}
	return p
}

func (p *BufferPool) Get() *bytes.Buffer {
	if p == nil {
		return &bytes.Buffer{}
	}
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func (p *BufferPool) Put(buf *bytes.Buffer) {
	if p == nil || buf == nil {
		return
	}
	if buf.Cap() > p.size*4 {
		return
	}
	buf.Reset()
	p.pool.Put(buf)
}
