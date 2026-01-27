package client

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/goccy/go-json"
	"rvpro3/radarvision.com/internal/tcphub"
	"rvpro3/radarvision.com/utils"
)

func TestVirtualRadar_Start(t *testing.T) {
	radar := &ClientRadar{}
	radar.OnUDPRead = onUDPRead
	radar.OnWriteFail = onWriteFail
	radar.OnWriteSuccess = onWriteSuccess
	radar.Start(utils.IP4Builder.FromString("192.168.0.4:40000"))
	//radar.Start(utils.IP4Builder.FromString("192.168.0.2:40000"))

	wg := sync.WaitGroup{}
	wg.Add(1)
	go sendAsRadar(radar, &wg)

	wg.Wait()
	radar.Stop()

	pretty, err := json.MarshalIndent(radar.Metrics, "", "  ")
	utils.Debug.Panic(err)
	fmt.Println(string(pretty))
}

func onWriteSuccess(v *ClientRadar, targetIP utils.IP4, dataOnly []byte) {
	fmt.Printf("OnWriteSuccess to %s with %d bytes\n", targetIP, len(dataOnly))
}

func onWriteFail(v *ClientRadar, targetIP utils.IP4, dataOnly []byte, err error) {
	fmt.Printf("OnWriteFail to %s with %d bytes, error %v\n", targetIP, len(dataOnly), err)
}

func sendAsRadar(radar *ClientRadar, wg *sync.WaitGroup) {
	var buffer [2 * utils.Kilobyte]byte

	packet := tcphub.PacketWrapper{}
	targetIP := utils.IP4Builder.FromString("192.168.0.2:39999")

	for n := 0; n < 5; n++ {
		packet.Init(buffer[:], 0, targetIP)
		packet.SetData([]byte("Hello World"))

		radar.Write(packet.GetPacket())
		time.Sleep(1 * time.Second)
	}
	wg.Done()
}

func onUDPRead(v *ClientRadar, addr utils.IP4, bytes []byte) {
	fmt.Printf("Response received from %s of buffer size %d\n", addr, len(bytes))
}
