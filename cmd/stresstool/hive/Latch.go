package hive

import "time"

type Latch struct {
	LatchNo   int
	Providers []*MessageProvider `xml:"Provider"`
	Time      time.Time          `xml:"-"`
	Index     int                `xml:"-"`
}

func (l *Latch) GetProvider() *MessageProvider {
	// When LatchNo is 0 then a single latch is expected
	if l.LatchNo == 0 {
		return l.Providers[0]
	}

	if l.Index == -1 {
		l.startLatch(0)
	}

	if l.isDone() {
		l.startLatch(l.Index + 1)
	}

	return l.Providers[l.Index]
}

func (l *Latch) startLatch(latchIndex int) {
	if latchIndex >= len(l.Providers) {
		latchIndex = 0
	}

	l.Time = time.Now()
	l.Index = latchIndex
}

func (l *Latch) isDone() bool {
	provider := l.Providers[l.Index]

	if provider.LatchDuration == 0 {
		if provider.IsEOF() {
			return true
		}
	} else {
		now := time.Now()
		diff := now.Sub(l.Time).Seconds()
		if int(diff) >= provider.LatchDuration {
			return true
		}
	}
	return false
}
