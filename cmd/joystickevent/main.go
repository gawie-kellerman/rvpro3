package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"rvpro3/radarvision.com/internal/hardware/m48"
	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/device/gpio"
)

var chips gpio.Chips
var counter int

func main() {
	chips.Init()
	defer chips.Close()

	attachEvent(int(m48.RightPort))
	attachEvent(int(m48.LeftPort))

	attachEvent(int(m48.UpPort))
	attachEvent(int(m48.DownPort))

	attachEvent(int(m48.EscapePort))
	attachEvent(int(m48.EnterPort))

	fmt.Println("Press any key on M48 Joystick")
	fmt.Println("Press Ctrl+C to quit.")
	time.Sleep(100 * time.Second)
}

func attachEvent(port int) {
	var err error
	var chip *gpio.Chip

	chip, err = chips.OpenByPort(port)
	utils.Debug.Panic(err)
	_, err = chip.ListenToLine(gpio.Util.GetOffset(port))
	chip.OnHandleGPIO = handler
	utils.Debug.Panic(err)

}

func handler(chip *gpio.Chip, event gpiocdev.LineEvent) {
	counter++

	var str strings.Builder
	str.Grow(100)
	str.WriteString(fmt.Sprintf("%03d - ", counter))

	portNo := gpio.Util.GetPortNo(chip.No, event.Offset)
	str.WriteString(m48.JoystickPort(portNo).String())

	switch event.Type {
	case gpiocdev.LineEventFallingEdge:
		str.WriteString("Pressed ")

	case gpiocdev.LineEventRisingEdge:
		str.WriteString("Released  ")
	}

	str.WriteString(", From:")
	str.WriteString(chip.Chip.Name)
	str.WriteString(" offset ")
	str.WriteString(strconv.Itoa(event.Offset))
	str.WriteString(" is port ")
	str.WriteString(strconv.Itoa(portNo))

	fmt.Printf("%s\n", str.String())
}
