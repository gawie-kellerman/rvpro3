package tcphub

import (
	"testing"
	"time"

	"rvpro3/radarvision.com/utils"
)

func TestHubHost_Start(t *testing.T) {
	GlobalHubHost.Start(utils.IP4Builder.FromOctets(127, 0, 0, 1, 45000))
	GlobalVirtualHost.Start(utils.IP4Builder.FromOctets(127, 0, 0, 1, 45000))

	go executeSendToClient()

	time.Sleep(1000 * time.Second)
}

func executeSendToClient() {
	targetIP := utils.IP4Builder.FromString("192.168.11.12:55555")
	packet := NewPacket([]byte("hello"))
	packet.TargetIP4 = targetIP.ToU32()
	packet.TargetPort = uint16(targetIP.Port)

	for {
		GlobalHubHost.WriteToClients(packet)
		time.Sleep(1 * time.Second)
	}
}
