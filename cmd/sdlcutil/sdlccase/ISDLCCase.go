package sdlccase

import (
	"rvpro3/radarvision.com/internal/sdlc/uartsdlc"
)

type ISDLCCase interface {
	SetService(*uartsdlc.SDLCService)

	Init()
	Start(func())
	Stop()
	SetOnTerminate(func(ISDLCCase))
}
