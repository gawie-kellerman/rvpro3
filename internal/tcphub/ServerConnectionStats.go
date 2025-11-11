package tcphub

import (
	"net"
	"time"

	"rvpro3/radarvision.com/utils"
)

type ServerConnectionStats struct {
	IPAddress utils.IP4

	ReadStats   readStats
	WriteStats  writeStats
	IsOpen      bool
	ConnectFrom time.Time
	ConnectTo   time.Time
}

type readStats struct {
	ErrCount   uint64
	ByteCount  uint64
	CycleCount uint64
}

type writeStats struct {
	/// When the packet size too big for the writeCache.  Packet is skipped
	OverflowCount uint64

	/// ServerConnection Connection Write failure... which closes the connection
	WriteErrCount uint64

	/// ServerConnection Connection Written only partially... does not close the connection
	PartialErrCount uint64

	WriteOKCount uint64
	WriteOKSize  uint64
}

func (stats *ServerConnectionStats) Start(addr net.Addr) {
	*stats = ServerConnectionStats{}
	stats.IPAddress = utils.IP4Builder.FromAddr(addr)
	stats.IsOpen = true
	stats.ConnectFrom = time.Now()
}

func (stats *ServerConnectionStats) Stop() {
	stats.ConnectTo = time.Now()
	stats.IsOpen = false
}
