package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"rvpro3/radarvision.com/internal/sdlc"
)

type RequestRepeater struct {
	RequestData []byte
	Cooldown    time.Duration
	Repeats     int
	Service     *sdlc.SDLCService
	OnTerminate func(*RequestRepeater)
}

func (rr *RequestRepeater) Start() {
	go rr.execute()
}

func (rr *RequestRepeater) execute() {
	for n := 0; n < rr.Repeats; n++ {
		fmt.Println("SDLC Request Repeat", hex.EncodeToString(rr.RequestData))
		rr.Service.Write(rr.RequestData)
		time.Sleep(rr.Cooldown)
	}

	if rr.OnTerminate != nil {
		rr.OnTerminate(rr)
	}
}
