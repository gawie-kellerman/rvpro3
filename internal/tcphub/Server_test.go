package tcphub

import (
	"fmt"
	"testing"
	"time"

	"rvpro3/radarvision.com/utils"
)

const iterations = 1000
const dataSize = 1400

func TestHubStartup(t *testing.T) {
	var testData [1500]byte
	copy(testData[dataSize-5:dataSize], "hello")
	//copy(testData[:], "hello1234567890")

	fmt.Println("Warming up for 15 seconds...")

	ip4 := utils.IP4Builder.FromPort(45000)
	GlobalTCPHubServer.Start(ip4)
	time.Sleep(15 * time.Second)

	for n := 0; n < iterations; n++ {
		packet1 := NewPacket(testData[:dataSize])
		GlobalTCPHubServer.SendToClients(packet1)
		//time.Sleep(time.Duration(1) * time.Second)
	}

	fmt.Printf("Sent %d packets with a street value of %d\n", iterations, iterations*(headerSize+dataSize))
	GlobalTCPHubServer.Stop()
}
