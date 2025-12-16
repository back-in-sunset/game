package cachex

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// TWindowK is a cache that supports top-k queries.
type TWindowK struct {
	cache   Cacher[int]
	sf      singleflight.Group
	mu      sync.RWMutex
	pending map[string]time.Time // key 最近一次查询的时间窗口
	KSec    int64
}

// NewTopK creates a new TopK cache with the given capacity.
func NewTopK(size int, KSec int64) *TWindowK {
	if KSec <= 0 {
		KSec = 5
	}

	return &TWindowK{
		cache:   NewClockCache[int](size),
		pending: make(map[string]time.Time),
		KSec:    KSec,
	}
}

// Get retrieves the value associated with the given key from the cache.
func (c *TWindowK) Get(key string, loader func() (int, error)) (int, error) {
	// 1. 先读缓存
	if val, ok := c.cache.Get(key); ok {
		return *val, nil
	}

	now := time.Now()

	// 2. 查看 5 秒合并窗口
	c.mu.RLock()
	deadline, ok := c.pending[key]
	c.mu.RUnlock()

	if ok && deadline.After(now) {
		// 已有请求正在飞行，等待结果
		v, err, _ := c.sf.Do(key, func() (any, error) {
			val, ok := c.cache.Get(key)
			if ok {
				return *val, nil
			}
			return 0, errors.New("request in progress")
		})

		return v.(int), err
	}

	// 3. 没有请求，则创建 5 秒窗口
	c.mu.Lock()
	c.pending[key] = now.Add(time.Second * time.Duration(c.KSec))
	c.mu.Unlock()

	// 4. 执行 singleflight：只会有 1 个实际 loader 调用
	val, err, _ := c.sf.Do(key, func() (any, error) {
		v, e := loader()
		if e == nil {
			c.cache.Set(key, v, 2*time.Second.Milliseconds()*c.KSec)
		}
		return v, e
	})

	return val.(int), err
}
