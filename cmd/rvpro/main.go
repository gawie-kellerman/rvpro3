package main

import (
	"encoding/json"
	"os"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/portbroker"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

func main() {
	udp := UdpExecutor{}
	udp.Start()

	time.Sleep(20 * time.Second)
	udp.Stop()
	udp.Save()
}

type UdpExecutor struct {
	aliveService service.UDPKeepAlive
	dataService  service.UDPData
	radars       portbroker.RadarChannels
}

func (u *UdpExecutor) Start() {
	ip4 := utils.IP4Builder.FromString("192.168.11.102:55555")

	u.radars.AttachTo(&u.dataService)
	u.radars.Start()
	u.aliveService.Init()
	u.aliveService.Start(ip4)
	u.dataService.Start(ip4)
}

func (u *UdpExecutor) Stop() {
	u.aliveService.Stop()
	u.dataService.Stop()
	u.radars.Stop()

	u.radars.AwaitStop()
}

func (u *UdpExecutor) Save() {
	type stats struct {
		Udp   service.UDPDataServiceStatistics
		Radar [4]portbroker.RadarStatistics
	}

	var toSave stats
	toSave.Udp = u.dataService.Metrics

	for index, radar := range u.radars.Radar {
		toSave.Radar[index] = radar.Stats
	}

	if bytes, err := json.Marshal(toSave); err != nil {
		panic(err)
	} else {
		if err = os.WriteFile("rvpro-stats.json", bytes, 0644); err != nil {
			panic(err)
		}
	}
}
