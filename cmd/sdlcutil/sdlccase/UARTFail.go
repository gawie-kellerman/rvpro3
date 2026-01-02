package sdlccase

import (
	"encoding/hex"
	"time"

	"rvpro3/radarvision.com/internal/sdlc/uartsdlc"
	"rvpro3/radarvision.com/utils"
)

type uartFailAction uint8

const MaxCyclesArg = "max-cycles"
const CycleDurationArg = "cycle-duration"
const DetectEveryArg = "detect-every"
const StatusEveryArg = "status-every"

type UARTFail struct {
	Mixin
	detectEnc     uartsdlc.SDLCRequestEncoder
	statusEnc     uartsdlc.SDLCRequestEncoder
	cycleCounter  int
	action        uartFailAction
	MaxCycles     int
	DetectEvery   int
	StatusEvery   int
	CycleDuration int
}

func (c *UARTFail) Init() {
	c.MaxCycles = utils.GlobalMap.Get(MaxCyclesArg).(int)
	c.DetectEvery = utils.GlobalMap.Get(DetectEveryArg).(int)
	c.StatusEvery = utils.GlobalMap.Get(StatusEveryArg).(int)
	c.CycleDuration = utils.GlobalMap.Get(CycleDurationArg).(int)

	utils.Print.Ln("Running UART Fail with:")
	utils.Print.Ln("  Max Cycles: ", c.MaxCycles)
	utils.Print.Ln("  Detect Every: ", c.DetectEvery, "cycles")
	utils.Print.Ln("  Status Every: ", c.StatusEvery, "cycles")
	utils.Print.Ln("  Metronome Duration: ", c.CycleDuration, "milliseconds")

}

func (c *UARTFail) Execute() {
	for !c.terminate {
		if c.cycleCounter%c.DetectEvery == 0 {
			c.sendDetect()
		}
		if c.cycleCounter%c.StatusEvery == 0 {
			c.sendStatusRequest()
		}

		if c.cycleCounter >= c.MaxCycles {
			c.terminate = true
		}

		time.Sleep(time.Duration(c.CycleDuration) * time.Millisecond)
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
