package main

import (
	"sync"

	"rvpro3/radarvision.com/internal/sdlc"
	"rvpro3/radarvision.com/utils"
)

type SDLCRequestRunner struct {
	Service  *sdlc.SDLCService
	Repeater *RequestRepeater
	Wg       sync.WaitGroup
}

func (rr *SDLCRequestRunner) Start() {
	rr.Service.OnTerminate = rr.onServiceTerminate
	rr.Repeater.OnTerminate = rr.onRepeaterTerminate

	rr.Wg.Add(1)
	rr.Service.Start()

	rr.Wg.Add(1)
	rr.Repeater.Start()
}

func (rr *SDLCRequestRunner) Await() {
	rr.Wg.Wait()
}

func (rr *SDLCRequestRunner) onServiceTerminate(service *sdlc.SDLCService) {
	utils.Print.Ln("Service terminated")
	rr.Wg.Done()
}

func (rr *SDLCRequestRunner) onRepeaterTerminate(repeater *RequestRepeater) {
	utils.Print.Ln("Repeater terminated")
	rr.Wg.Done()
	rr.Service.Stop()
}
