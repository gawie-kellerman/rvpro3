package broker

import (
	"sync"
	"time"

	"rvpro3/radarvision.com/utils"
)

var messagePool = sync.Pool{
	New: func() interface{} {
		return new(UDPMessage)
	},
}

type UDPMessage struct {
	CreateOn  time.Time
	IPAddress utils.IP4
	Buffer    [4000]byte
	BufferLen int
}
