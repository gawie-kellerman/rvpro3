package tcphub

import (
	"net"
	"time"

	"rvpro3/radarvision.com/utils"
)

type HubClientStat struct {
	ClientIP     uint32
	ConnectAt    int64
	DisconnectAt int64
	WriteCount   uint32
	WriteSize    uint32
	WriteAt      int64
	ReadCount    uint32
	ReadSize     uint32
	ReadAt       int64
}

func (c *HubClientStat) RegisterConnect(remoteAddr net.Addr) {
	ip4 := utils.IP4Builder.FromAddr(remoteAddr)
	c.ClientIP = ip4.ToU32()
	c.ConnectAt = time.Now().UnixMilli()
	c.DisconnectAt = 0
	c.WriteCount = 0
	c.WriteSize = 0
	c.WriteAt = 0
	c.ReadCount = 0
	c.ReadSize = 0
	c.ReadAt = 0
}

func (c *HubClientStat) RegisterRead(read int) {
	c.ReadCount++
	c.ReadSize += uint32(read)
	c.ReadAt += time.Now().UnixMilli()
}

func (c *HubClientStat) RegisterWrite(write int) {
	c.WriteCount++
	c.WriteSize += uint32(write)
	c.WriteAt = time.Now().UnixMilli()
}

func (c *HubClientStat) RegisterDisconnect() {
	c.DisconnectAt = time.Now().UnixMilli()
}
