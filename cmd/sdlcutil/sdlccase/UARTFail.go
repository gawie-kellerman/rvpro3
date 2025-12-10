package sdlccase

import (
	"encoding/hex"
	"time"

	"rvpro3/radarvision.com/internal/sdlc"
	"rvpro3/radarvision.com/utils"
)

type uartFailAction uint8

const MaxCyclesArg = "max-cycles"
const DetectEveryArg = "detect-every"
const StatusEveryArg = "status-every"

type UARTFail struct {
	Mixin
	detectEnc    sdlc.SDLCRequestEncoder
	statusEnc    sdlc.SDLCRequestEncoder
	cycleCounter int
	action       uartFailAction
	MaxCycles    int
	DetectEvery  int
	StatusEvery  int
}

func (c *UARTFail) Init() {
	c.MaxCycles = utils.Args.Get(MaxCyclesArg).(int)
	c.DetectEvery = utils.Args.Get(DetectEveryArg).(int)
	c.StatusEvery = utils.Args.Get(StatusEveryArg).(int)

	utils.Print.Ln("Running UART Fail with:")
	utils.Print.Ln("  Max Cycles: ", c.MaxCycles)
	utils.Print.Ln("  Detect Every: ", c.DetectEvery, "seconds")
	utils.Print.Ln("  Status Every: ", c.StatusEvery, "seconds")
}

func (c *UARTFail) Execute() {
	for !c.terminate {
		if c.cycleCounter%c.DetectEvery == 0 {
			c.sendDetect()
		} else if c.cycleCounter%c.StatusEvery == 0 {
			c.sendStatusRequest()
		}

		if c.cycleCounter >= c.MaxCycles {
			c.terminate = true
		}

		time.Sleep(1 * time.Second)
		c.cycleCounter++
	}

	c.terminated = true

	if c.onTerminate != nil {
		c.onTerminate(c)
	}
}

func (c *UARTFail) sendStatusRequest() {
	data, err := c.statusEnc.DynamicStatus()
	utils.Debug.Panic(err)

	utils.Print.Ln("Sending Status Request: ", hex.EncodeToString(data))
	c.service.Write(data)
}

func (c *UARTFail) sendDetect() {
	data, err := c.detectEnc.TS2Detect(0)
	utils.Debug.Panic(err)

	utils.Print.Ln("Sending Detect Request: ", hex.EncodeToString(data))
	c.service.Write(data)
}
