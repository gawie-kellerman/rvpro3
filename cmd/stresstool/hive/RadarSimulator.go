package hive

import (
	"fmt"
	"time"

	"rvpro3/radarvision.com/cmd/stresstool/config"
	"rvpro3/radarvision.com/utils"
)

type RadarSimulator struct {
	connection utils.UDPClientConnection
	terminate  bool
	terminated bool
	providers  *MessageProviders
	CooldownMs int
	RadarIP    utils.IP4
	RVProIP    utils.IP4
}

type RadarSimulatorStats struct {
	RVProIP string   `xml:"rvproIP,attr"`
	RadarIP string   `xml:"radarIP,attr"`
	Latches []*Latch `xml:"Latch"`
}

func (r *RadarSimulator) GetStats() *RadarSimulatorStats {
	result := new(RadarSimulatorStats)
	result.Latches = r.providers.Latches
	result.RVProIP = r.RVProIP.String()
	result.RadarIP = r.RadarIP.String()
	return result
}

func (r *RadarSimulator) Init(localIPAddr utils.IP4, targetIP utils.IP4, types []config.MessageType, cooldownMs int) {
	r.RadarIP = localIPAddr
	r.RVProIP = targetIP
	r.CooldownMs = cooldownMs
	r.connection.Init(localIPAddr, targetIP, r, 3)
	r.providers = new(MessageProviders)
	r.providers.Init(types)
}

func (r *RadarSimulator) Start() {
	go r.execute()
}

func (r *RadarSimulator) execute() {
	for !r.terminate {
		if r.connection.Connect() {
			cnx := r.connection.GetConnection()

			for _, latch := range r.providers.Latches {
				sendFile := latch.GetProvider().GetNextFile()

				if sendFile == nil {
					continue
				}

				stats := &sendFile.Stats
				stats.Iterations++

				now := time.Now()
				bytesWritten, err := cnx.Write(sendFile.Bytes())
				since := time.Since(now)

				if err != nil {
					fmt.Printf("Error on WriteToUDP: %s, %d\n", err, bytesWritten)
					stats.SendErrs++
					r.connection.Disconnect()
				} else {
					stats.SendBytes += uint64(bytesWritten)
					stats.AddSendCount(uint64(bytesWritten), since)
				}
			}
		}

		time.Sleep(time.Duration(r.CooldownMs) * time.Millisecond)
	}

	r.terminated = true
}

func (r *RadarSimulator) Stop() {
	r.terminate = true
}

func (r *RadarSimulator) AwaitStop() {
	for !r.terminated {
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}
