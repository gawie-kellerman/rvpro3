package client

import (
	"testing"
	"time"

	"rvpro3/radarvision.com/utils"
)

func TestHubClient_Start(t *testing.T) {
	hc := Client{}
	hc.Start(utils.IP4Builder.FromString("192.168.0.2:35000"))
	time.Sleep(10 * time.Minute)
	hc.StopAndJoin()
}
