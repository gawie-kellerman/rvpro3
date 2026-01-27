package server

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"rvpro3/radarvision.com/internal/router/client"
	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

func TestServer_Start(t *testing.T) {
	wg := sync.WaitGroup{}
	writer := DummyUDPPacketWriter{}
	hs := Server{}
	hs.Start(utils.IP4Builder.FromString("192.168.11.1:45000"), &writer)

	hc := startClient(utils.IP4Builder.FromString("192.168.11.1:45000"))
	wg.Add(1)

	go writeViaServer(utils.IP4Builder.FromString("192.168.11.12:55555"), &hs, &wg)
	go writeViaServer(utils.IP4Builder.FromString("192.168.11.13:55555"), &hs, &wg)
	//go writeViaServer(utils.IP4Builder.FromString("192.168.11.14:55555"), &hs, &wg)
	//go writeViaServer(utils.IP4Builder.FromString("192.168.11.15:55555"), &hs, &wg)
	wg.Wait()

	hc.StopAndJoin()
}

func startClient(ip4 utils.IP4) client.Client {
	hc := client.Client{}
	hc.Start(ip4)
	return hc
}

func writeViaServer(radarIP utils.IP4, server *Server, wg *sync.WaitGroup) {
	wg.Add(1)
	var buffer [2048]byte

	unitIP := utils.IP4Builder.FromString("192.168.11.1:55555")

	for i := 0; i < 100000; i++ {
		packet := tcphub.PacketWrapper{}
		packet.Init(buffer[:], 0, unitIP, radarIP)
		packet.SetData([]byte(fmt.Sprintf("<<Iteration %003d>>, <<%s>>", i, strings.Repeat("1", i%30))))
		server.Write(packet.GetPacket())
		time.Sleep(1 * time.Second)
	}

	wg.Done()
}

type DummyUDPPacketWriter struct {
}

func (d *DummyUDPPacketWriter) WritePacket(packetData []byte) error {
	packet := tcphub.PacketWrapper{
		Buffer: packetData,
	}

	fmt.Printf("target ip: %s,  data size: %d data: %s\n", packet.GetTargetIP4(), packet.GetDataSize(), string(packet.GetData()))
	return nil
}
