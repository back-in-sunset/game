package bytes

import (
	"testing"
)

func TestBuffer(t *testing.T) {
	p := NewPool(2, 10)

	t.Errorf("p.free=%p", p.free)
	b1 := p.Get()
	t.Errorf("p.free=%p b1=%p", p.free, b1)
	if b1.Bytes() == nil || len(b1.Bytes()) == 0 {
		t.FailNow()
	}
	b2 := p.Get()
	t.Errorf("p.free=%p b2=%p", p.free, b2)
	if b2.Bytes() == nil || len(b2.Bytes()) == 0 {
		t.FailNow()
	}

	b3 := p.Get()
	t.Errorf("p.free=%p b2.next=%p b3=%p", p.free, b2.next, b3)
	if b3.Bytes() == nil || len(b3.Bytes()) == 0 {
		t.FailNow()
	}
	p.Put(b3)
	t.Errorf("p.free=%p b3=%p", p.free, b3)

	p.Put(b2)
	t.Errorf("p.free=%p b2=%p", p.free, b2)

	p.Put(b1)
	t.Errorf("p.free=%p b1=%p", p.free, b1)
}
