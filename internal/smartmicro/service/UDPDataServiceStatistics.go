package service

import (
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/common"
)

type udpDataStatistic int

const (
	udpHead udpDataStatistic = iota + 100
	UdpIncorrectRadar
	udpSocketSuccess
	updSocketFailure
	udpDataReceived
	updDataNotReceived
	udpDataTotal
	udpTail
)

type UDPDataServiceStatistics struct {
	Data [udpTail - udpHead - 1]common.CounterStatistic
}

func (s *UDPDataServiceStatistics) Init() {
	for i := 0; i < len(s.Data); i++ {
		data := &s.Data[i]
		data.Id = int(udpHead) + i + 1
	}
}

func (s *UDPDataServiceStatistics) Register(stat udpDataStatistic, now time.Time) bool {
	data := &s.Data[stat-udpHead-1]

	if data.Count == 0 {
		data.Count = 1
		data.FirstOn = now.Unix()
		data.LastOn = data.FirstOn
		return true
	}

	data.Count = data.Count + 1
	data.LastOn = now.Unix()
	return false
}

func (s *UDPDataServiceStatistics) Aggregate(stat udpDataStatistic, count uint64, now time.Time) bool {
	data := &s.Data[stat-udpHead-1]

	if data.Count == 0 {
		data.Count = 1
		data.FirstOn = now.Unix()
		data.LastOn = data.FirstOn
		return true
	}

	data.Count = data.Count + count
	data.LastOn = now.Unix()
	return false
}
