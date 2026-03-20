package main

import (
	"fmt"
	"time"

	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/device/gpio"
)

const rightPort = 110

func main() {
	fl := gpio.File{}
	fl.Init(gpio.DirectionIn, rightPort)
	utils.Debug.Panic(fl.Open())

	duration := time.Duration(0)
	for i := range 10000 {
		now := time.Now()
		v, err := fl.Read()
		then := time.Now()

		duration = then.Sub(now)

		utils.Debug.Panic(err)
		fmt.Println(i, "value", v)
		//in, err := c.LineInfo(leftPort)
		//utils.Debug.Panic(err)
		//fmt.Println("Iteration", i)
		//fmt.Println("Config.Direction", in.Config.Direction)
		//fmt.Println("Config.Bias", in.Config.Bias)
		//fmt.Println("Config.Debounced", in.Config.Debounced)
		//fmt.Println("Config.ActiveLow", in.Config.ActiveLow)
		//fmt.Println("Config.DebouncePeriod", in.Config.DebouncePeriod)
		//fmt.Println("Config.EdgeDetection", in.Config.EdgeDetection)
		//fmt.Println("Config.EventClock", in.Config.EventClock)
		//fmt.Println("Name", in.Name, "used", in.Used, "offset", in.GetOffset, "consumer", in.Consumer)
		//time.Sleep(1000 * time.Millisecond)
	}
	fmt.Println(duration.Nanoseconds())
}
