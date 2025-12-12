package hive

import (
	"math"
	"time"
)

type MessageFileStats struct {
	Iterations      uint
	SendCount       uint
	SendBytes       uint64
	SendErrs        uint64
	MaxSendTimeNs   time.Duration
	MinSendTimeNs   time.Duration
	TotalSendTimeNs uint64
}

func (s *MessageFileStats) AddSendCount(bytes uint64, since time.Duration) {
	s.SendCount += 1
	s.SendBytes += bytes

	s.MaxSendTimeNs = max(s.MaxSendTimeNs, since)
	s.MinSendTimeNs = min(s.MinSendTimeNs, since)
	s.TotalSendTimeNs += uint64(since.Nanoseconds())
}

func (s *MessageFileStats) Init() {
	s.MinSendTimeNs = math.MaxInt64
	s.MaxSendTimeNs = math.MinInt64
}
