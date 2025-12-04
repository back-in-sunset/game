package idx

import (
	"sync"
	"time"
)

const (
	sequenceBits = 12
	sequenceMask = (1 << sequenceBits) - 1
)

type Snowflake struct {
	mu        sync.Mutex
	nodeID    int64
	sequence  int64
	lastStamp int64

	pool chan int64
	time func() int64
}

func NewSnowflakeWithPool(nodeID int64, poolSize int, batchSize int, time func() int64) *Snowflake {
	s := &Snowflake{
		nodeID: nodeID,
		pool:   make(chan int64, poolSize),
		time:   time,
	}

	// 启动后台 refill
	go s.refill(batchSize)

	return s
}

// 生成一个 ID（仅内部使用）
func (s *Snowflake) nextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.time()
	if now < s.lastStamp {
		now = s.lastStamp
	}

	if now == s.lastStamp {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now <= s.lastStamp {
				now = s.time()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastStamp = now

	t := (now - 1609459200000) << (sequenceBits + 10)
	n := s.nodeID << sequenceBits
	return t | n | s.sequence
}

// 后台批量 refill ID
func (s *Snowflake) refill(batch int) {
	for {
		needed := cap(s.pool) - len(s.pool)
		if needed <= 0 {
			time.Sleep(time.Microsecond * 50)
			continue
		}

		// 本次最多填充 batch 个
		n := batch
		if n > needed {
			n = needed
		}

		for i := 0; i < n; i++ {
			s.pool <- s.nextID()
		}
	}
}

// ----------------------------------------
// 对外接口
// ----------------------------------------

// Next 生成下一个ID
func (s *Snowflake) Next() int64 {
	return <-s.pool
}

// NextN 批量生成 n 个 ID
func (s *Snowflake) NextN(n int) []int64 {
	res := make([]int64, n)
	for i := 0; i < n; i++ {
		res[i] = <-s.pool
	}
	return res
}
