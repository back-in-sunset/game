package cachex

import (
	"sync"
	"time"
)


type entry[T any] struct {
	key        string
	value      T
	expAt      int64
	referenced bool
}

// ClockCache implements a fixed-size in-memory cache with CLOCK eviction policy.
type ClockCache[T any] struct {
	capacity int
	slots    []*entry[T]
	clockPtr int
	index    map[string]int
	mu       sync.Mutex
}

// NewClockCache creates a new ClockCache with the given capacity.
func NewClockCache[T any](cap int) *ClockCache[T] {
	return &ClockCache[T]{
		capacity: cap,
		slots:    make([]*entry[T], cap),
		index:    make(map[string]int),
	}
}

func now() int64 {
	return time.Now().UnixMilli()
}

// Get retrieves a value from the cache by key.
func (c *ClockCache[T]) Get(key string) (*T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pos, ok := c.index[key]
	if !ok {
		return new(T), false
	}
	e := c.slots[pos]

	if e.expAt > 0 && e.expAt <= now() {
		// expired
		c.removeSlot(pos)
		return nil, false
	}

	e.referenced = true
	return &e.value, true
}

func (c *ClockCache[T]) removeSlot(pos int) {
	e := c.slots[pos]
	if e != nil {
		delete(c.index, e.key)
	}
	c.slots[pos] = nil
}

// Set adds or updates a value in the cache with an optional TTL in milliseconds.
func (c *ClockCache[T]) Set(key string, value T, ttlMs int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// update existing
	if pos, ok := c.index[key]; ok {
		e := c.slots[pos]
		e.value = value
		e.referenced = true
		if ttlMs > 0 {
			e.expAt = now() + ttlMs
		}
		return
	}

	// insert new
	pos := c.findVictim()
	c.removeSlot(pos)

	expAt := int64(0)
	if ttlMs > 0 {
		expAt = now() + ttlMs
	}

	c.slots[pos] = &entry[T]{
		key:        key,
		value:      value,
		expAt:      expAt,
		referenced: true,
	}
	c.index[key] = pos
}

func (c *ClockCache[T]) findVictim() int {
	for {
		e := c.slots[c.clockPtr]

		// empty slot
		if e == nil {
			pos := c.clockPtr
			c.clockPtr = (c.clockPtr + 1) % c.capacity
			return pos
		}

		// expired
		if e.expAt > 0 && e.expAt <= now() {
			pos := c.clockPtr
			c.clockPtr = (c.clockPtr + 1) % c.capacity
			return pos
		}

		// referenced? give second chance
		if e.referenced {
			e.referenced = false
			c.clockPtr = (c.clockPtr + 1) % c.capacity
			continue
		}

		// unreferenced victim
		pos := c.clockPtr
		c.clockPtr = (c.clockPtr + 1) % c.capacity
		return pos
	}
}

// Delete removes a key from the cache.
func (c *ClockCache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if pos, ok := c.index[key]; ok {
		c.removeAt(pos)
	}
}

func (c *ClockCache[T]) removeAt(pos int) {
	e := c.slots[pos]
	if e != nil {
		delete(c.index, e.key)
	}
	c.slots[pos] = nil
}
