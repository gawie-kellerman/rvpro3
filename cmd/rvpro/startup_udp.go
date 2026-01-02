package main

//func doUDPStartup() {
//	exe := udpExecutor{}
//	exe.Run()
//}
//
//type udpExecutor struct {
//	alive        service.UDPKeepAlive
//	data         service.UDPData
//	channels     udp.RadarChannels
//	sdlc         *uartsdlc.SDLCService
//	sdlcExecutor *uartsdlc.SDLCExecutorService
//	workflows    interfaces.IUDPWorkflowBuilder
//}
//
//func (u *udpExecutor) Run() {
//	Executor = u
//	u.workflows = workflows.WorkflowBuilder{}
//
//	setup := config.RVProSetup{}
//	setup.UDPKeepAlive(&u.alive)
//	setup.UDPData(&u.data)
//	setup.Channels(&u.channels, &u.data)
//	setup.ZeroLog()
//
//	u.alive.Start()
//	u.data.Start()
//	u.channels.Start(u.workflows)
//	u.startSDLC(&setup)
//
//	startupWeb()
//
//	u.channels.AwaitStop(3000 * time.Millisecond)
//}
//
//func (u *udpExecutor) Stop() {
//	u.alive.Stop()
//	u.data.Stop()
//	u.channels.Stop()
//	u.channels.AwaitStop(100 * time.Millisecond)
//}
//
//func (u *udpExecutor) StopRadars() {
//	u.alive.Stop()
//	u.data.Stop()
//}
//
//func (u *udpExecutor) StartRadars() {
//	u.alive.Start()
//	u.data.Start()
//}
//
//func (u *udpExecutor) IsRadarsStopped() bool {
//	return u.alive.IsTerminated() && u.data.IsTerminated()
//}
//
//func (u *udpExecutor) startSDLC(setup *config.RVProSetup) {
//	switch runtime.GOOS {
//	case "linux":
//		u.sdlc = utils.GlobalState.
//			Set(uartsdlc.SDLCServiceStateName, new(uartsdlc.SDLCService)).(*uartsdlc.SDLCService)
//
//		setup.SDLCUARTService(u.sdlc)
//
//		u.sdlcExecutor = new(uartsdlc.SDLCExecutorService)
//		u.sdlcExecutor.Start()
//		u.sdlc.OnReadMessage = u.sdlcExecutor.OnReadMessage
//
//		u.sdlc.Start()
//	default:
//	}
//}
//
//var Executor IExecutor
//
//type IExecutor interface {
//	IsRadarsStopped() bool
//	StopRadars()
//	StartRadars()
//}
