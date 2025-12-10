package sdlccase

import "rvpro3/radarvision.com/internal/sdlc"

type ISDLCCase interface {
	SetService(*sdlc.SDLCService)

	Init()
	Start(func())
	Stop()
	SetOnTerminate(func(ISDLCCase))
}
