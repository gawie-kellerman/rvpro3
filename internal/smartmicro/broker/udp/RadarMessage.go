package udp

import (
	"sync"
	"time"

	"rvpro3/radarvision.com/utils"
)

var messagePool = sync.Pool{
	New: func() interface{} {
		return new(RadarMessage)
	},
}

type RadarMessage struct {
	CreateOn  time.Time
	IPAddress utils.IP4
	Buffer    [1500]byte
	BufferLen int
}
