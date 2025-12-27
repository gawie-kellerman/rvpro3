package main

import (
	"time"

	"rvpro3/radarvision.com/internal/config"
	"rvpro3/radarvision.com/internal/smartmicro/broker/udp"
	"rvpro3/radarvision.com/internal/smartmicro/interfaces"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/internal/smartmicro/workflows/udp/common"
)

func doUDPStartup() {
	exe := udpExecutor{}
	exe.Run()
}

type udpExecutor struct {
	alive     service.UDPKeepAlive
	data      service.UDPData
	channels  udp.RadarChannels
	workflows interfaces.IUDPWorkflowBuilder
}

func (u *udpExecutor) Run() {
	u.workflows = common.WorkflowBuilder{}

	setup := config.RVProSetup{}
	setup.UDPKeepAlive(&u.alive)
	setup.UDPData(&u.data)
	setup.Channels(&u.channels, &u.data)
	setup.ZeroLog()

	u.alive.Start()
	u.data.Start()
	u.channels.Start(u.workflows)

	startupWeb()

	u.channels.AwaitStop(3000 * time.Millisecond)
}

func (u *udpExecutor) Stop() {
	u.alive.Stop()
	u.data.Stop()
	u.channels.Stop()
	u.channels.AwaitStop(100 * time.Millisecond)
}
