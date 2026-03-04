package websdlc

import (
	"math"
	"testing"

	"rvpro3/radarvision.com/utils"
)

const basePath = "http://192.168.0.80"

func TestSdlcWebApi_GetStatus4(t *testing.T) {
	status, err := SDLCWebApi.GetStatus4()
	utils.Debug.Panic(err)

	utils.Print.Ln(status)
}

func TestSdlcWebApi_GetStatus5(t *testing.T) {
	res, err := SDLCWebApi.SendTS2Detect(15, math.MaxUint64)
	utils.Debug.Panic(err)

	utils.Print.Ln(res)
}
