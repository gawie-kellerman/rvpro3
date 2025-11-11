package main

import (
	"fmt"

	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

func main() {
	serialParser := service.SerialDataParser{}

	serialService := service.SerialDataService{
		BaudRate: 921600,
		PortName: "/dev/ttyUSB0",
		OnData: func(service *service.SerialDataService, data []byte) []byte {
			return serialParser.Parse(data)
		},
	}

	serialService.LoopGuard = &utils.InfiniteLoopGuard{}
	serialService.RetryGuard.ModCycles = 3
	serialService.OnError = func(_ any, err error) {
		fmt.Println(err)
	}

	serialService.Execute()

	//var buffer [32000]byte
	//
	//mode := &serial.Mode{
	//	BaudRate: 921600,
	//}
	//
	//fmt.Println("opening serial port")
	//
	//port, err := serial.Listen("/dev/ttyUSB0", mode)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//for {
	//	fmt.Println("reading from serial...")
	//	read, err := port.ReadPortData(buffer[:])
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	fmt.Println("read ", read, " bytes from serial")
	//}

}
