package main

import (
	"fmt"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"rvpro3/radarvision.com/utils"
)

const leftPort = 5
const rightPort = 110
const upPort = 1
const downPort = 112
const enterPort = 4
const escapePort = 9

func main() {
	cc := gpiocdev.Chips()

	fmt.Printf("Chips: %v\n", cc)
	//for chip := range cc {
	//	fmt.Println(chip)
	//}

	c, err := gpiocdev.NewChip("gpiochip0")
	utils.Debug.Panic(err)
	defer c.Close()

	line, err := c.RequestLine(leftPort, gpiocdev.LineEdgeBoth)
	utils.Debug.Panic(err)

	duration := time.Duration(0)
	for i := range 10000 {
		now := time.Now()
		v, err := line.Value()
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
