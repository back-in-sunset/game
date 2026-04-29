package verify

import (
	"sync"
	"time"
)

type CodeStore interface {
	Set(target string, code string, ttl time.Duration)
	VerifyAndDelete(target string, code string) bool
}

type memoryCodeStore struct {
	mu    sync.Mutex
	items map[string]codeItem
}

type codeItem struct {
	code   string
	expire time.Time
}

func NewMemoryCodeStore() CodeStore {
	return &memoryCodeStore{
		items: make(map[string]codeItem),
	}
}

func (s *memoryCodeStore) Set(target string, code string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[target] = codeItem{
		code:   code,
		expire: time.Now().Add(ttl),
	}
}

func (s *memoryCodeStore) VerifyAndDelete(target string, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[target]
	if !ok {
		return false
	}
	if time.Now().After(item.expire) {
		delete(s.items, target)
		return false
	}
	if item.code != code {
		return false
	}
	delete(s.items, target)
	return true
}
