package main

import (
	"time"

	"rvpro3/radarvision.com/internal/router/client"
	"rvpro3/radarvision.com/utils"
)

func main() {
	hc := client.RouterClient{}
	hc.Start(utils.IP4Builder.FromString("192.168.0.103:45001"))
	time.Sleep(6000 * time.Minute)
	hc.StopAndJoin()
}
