package portbroker

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
	Buffer    [3000]byte
	BufferLen int
}
