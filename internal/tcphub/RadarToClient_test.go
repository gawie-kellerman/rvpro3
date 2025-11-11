package tcphub

import (
	"fmt"
	"testing"
	"time"

	"rvpro3/radarvision.com/utils"
)

var isOpen = false

const data = "abcdefghij"

// TestRadarToClient test must have the virtual network (.12 .. .15 running)
func TestRadarToClient(t *testing.T) {
	// Startup the server

	startTCPHub()

	time.Sleep(100 * time.Second)
}

func TestStopStart(t *testing.T) {
	startTCPHub()
	startTCPHub()
	startTCPHub()
}

func startTCPHub() {
	isOpen = false
	fmt.Println("Starting TCP Hub")
	serverIP := utils.IP4Builder.FromString("127.0.0.1:45000")
	clientIP := utils.IP4Builder.FromString("127.0.0.1:55555")

	GlobalTCPHubServer.Start(serverIP)

	GlobalTCPHubClient.OnOpenConnection = func(sender any) {
		if !isOpen {
			go sendToClientsAbort()
			isOpen = true
		}
	}
	time.Sleep(1 * time.Second)

	GlobalTCPHubClient.OnCloseConnection = func(sender any) {
		fmt.Println("GlobalTCPHubClient OnCloseConnection")
	}
	GlobalTCPHubClient.Start(serverIP, clientIP)

	GlobalTCPHubClient.OnError = func(sender any, err error) {
		fmt.Println("GlobalTCPHubClient OnError", err)
	}

	time.Sleep(15 * time.Second)
	fmt.Println("Stopping TCP Hub")
	GlobalTCPHubClient.Stop()
	fmt.Println("got here")
	GlobalTCPHubServer.Stop()
}

func sendToClients() {
	time.Sleep(5 * time.Second)
	for n := 0; n < 1000; n++ {
		radarIndex := n % 4
		radarIP := utils.RadarIPOf(radarIndex)
		packet := makePacket(radarIP)
		GlobalTCPHubServer.SendToClients(packet)
		time.Sleep(10 * time.Millisecond)
	}
}

func sendToClientsAbort() {
	time.Sleep(5 * time.Second)
	for n := 0; n < 1000; n++ {
		radarIndex := n % 4
		radarIP := utils.RadarIPOf(radarIndex)
		packet := makePacket(radarIP)
		GlobalTCPHubServer.SendToClients(packet)
		time.Sleep(10 * time.Millisecond)

		//if n == 500 {
		//	GlobalTCPHubClient.Stop()
		//	for !GlobalTCPHubClient.Terminated {
		//		time.Sleep(1 * time.Second)
		//	}
		//	serverIP := utils.IP4Builder.FromString("127.0.0.1:45000")
		//	clientIP := utils.IP4Builder.FromString("127.0.0.1:55555")
		//	GlobalTCPHubClient.Start(serverIP, clientIP)
		//}
	}
}

func makePacket(ip utils.IP4) Packet {
	res := NewPacket([]byte(data))
	res.TargetIP4 = ip.ToU32()
	res.TargetPort = uint16(ip.Port)
	res.Type = PtUdpForward
	return res
}
