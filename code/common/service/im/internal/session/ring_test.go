package session

import "testing"

func TestRingRoundsToPowerOfTwo(t *testing.T) {
	r := NewRing(3)
	if got, want := r.Cap(), 4; got != want {
		t.Fatalf("Cap() = %d, want %d", got, want)
	}
}

func TestRingForcePushDropsOldest(t *testing.T) {
	r := NewRing(2)
	if err := r.Push([]byte("a")); err != nil {
		t.Fatalf("Push(a) error = %v", err)
	}
	if err := r.Push([]byte("b")); err != nil {
		t.Fatalf("Push(b) error = %v", err)
	}
	dropped := r.ForcePush([]byte("c"))
	if string(dropped) != "a" {
		t.Fatalf("ForcePush() dropped = %q, want a", string(dropped))
	}
	first, _ := r.Pop()
	second, _ := r.Pop()
	if string(first) != "b" || string(second) != "c" {
		t.Fatalf("ring order = %q,%q want b,c", string(first), string(second))
	}
}
