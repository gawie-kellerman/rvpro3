package sdlccase

import (
	"time"

	"rvpro3/radarvision.com/internal/sdlc"
)

type Mixin struct {
	service     *sdlc.SDLCService
	terminate   bool
	terminated  bool
	onTerminate func(ISDLCCase)
}

func (c *Mixin) SetService(service *sdlc.SDLCService) {
	c.service = service
}

func (c *Mixin) Start(execute func()) {
	c.terminate = false
	c.terminated = false
	go execute()
}

func (c *Mixin) Stop() {
	c.terminate = true

	for !c.terminated {
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *Mixin) SetOnTerminate(onTerminate func(ISDLCCase)) {
	c.onTerminate = onTerminate
}

func (c *Mixin) execute() {
	c.terminate = true

	for !c.terminate {

	}

	c.terminated = true
}
