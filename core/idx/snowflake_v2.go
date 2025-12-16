package idx

import (
	"sync"
	"time"
)

const (
	sequenceBits = 12
	sequenceMask = (1 << sequenceBits) - 1
)

// Snowflake 雪花算法实现
type Snowflake struct {
	mu        sync.Mutex
	nodeID    int64
	sequence  int64
	lastStamp int64

	pool     chan int64
	timeFunc func() int64
	stopChan chan struct{} // 1. 新增：用于优雅停止后台循环
	wg       sync.WaitGroup
}

// NewSnowflakeWithPool 创建一个新的 Snowflake 实例，同时启动后台 goroutine 填充 ID 池
// nodeID: 节点ID，范围为 [0, MaxNodeID)
// poolSize: ID 池大小，建议设为 batchSize 的倍数
// batchSize: 每次填充的 ID 数量，建议设为 1000 左右
// timeFunc: 时间函数，用于获取当前时间戳（毫秒级）
func NewSnowflakeWithPool(nodeID int64, poolSize int, batchSize int, timeFunc func() int64) *Snowflake {
	s := &Snowflake{
		nodeID:   nodeID,
		pool:     make(chan int64, poolSize),
		timeFunc: timeFunc,
		stopChan: make(chan struct{}),
	}

	s.wg.Add(1)
	// 启动后台 refill
	go s.refill(batchSize)

	return s
}

// Close 停止后台 goroutine，释放资源
func (s *Snowflake) Close() {
	close(s.stopChan) // 通知后台goroutine退出
	s.wg.Wait()       // 等待其安全退出
}

// 生成一个 ID（仅内部使用）
func (s *Snowflake) nextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.timeFunc()
	if now < s.lastStamp {
		now = s.lastStamp
	}

	if now == s.lastStamp {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			// 关键修复1：避免忙等，等待下一个毫秒的到来
			// 计算需要等待的时间（转换为纳秒），然后让出CPU
			nextMs := s.lastStamp + 1
			// 注意：这里假设 timeFunc 返回毫秒时间戳
			// 如果timeFunc是自定义的，需要根据其实现调整等待逻辑
			for now < nextMs {
				now = s.lastStamp + 1
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
	defer s.wg.Done() // 保证在退出时通知WaitGroup

	ticker := time.NewTicker(10 * time.Millisecond) // 关键修复2：使用固定间隔的Ticker，而非忙等
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			// 收到停止信号，退出goroutine
			return
		case <-ticker.C:
			// 每隔一段时间检查并填充一次
			s.tryRefill(batch)
		}
	}
}

// 解耦出来的填充逻辑
func (s *Snowflake) tryRefill(batch int) {
	needed := cap(s.pool) - len(s.pool)
	if needed <= 0 {
		return // 通道已满，直接返回，无需任何休眠或循环
	}

	// 本次最多填充 batch 个
	n := batch
	if n > needed {
		n = needed
	}
	for i := 0; i < n; i++ {
		// 使用select避免在池满时阻塞，增加健壮性
		select {
		case s.pool <- s.nextID():
			// 成功写入
		default:
			// 通道意外已满，直接退出填充循环
			return
		}
	}
}

// ----------------------------------------
// 对外接口 (保持不变)
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
