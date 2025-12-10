package sdlc

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"rvpro3/radarvision.com/utils"
)

func Test_CMUFrameStream(t *testing.T) {
	decoder := SDLCResponseDecoder{}
	utils.Debug.Panic(decoder.InitFromHex("024100000000220109CA8303"))

	assert.Equal(t, CMUFrameStreamCode, decoder.GetIdentifier())
	cmuFrame, err := decoder.GetCMUFrame()
	utils.Debug.Panic(err)

	fmt.Printf("%s\n", cmuFrame.String())
}

func Test_DateTimeStream(t *testing.T) {
	decodeDate(t, "0242180a7d22173b39482703")
	decodeDate(t, "0242180a7d22173b30d90e03")
	decodeDate(t, "0242180a7d22000000f0c103")
	decodeDate(t, "0242180a7d2200001d335d03")
}

func decodeDate(t *testing.T, hexStr string) {
	decoder := SDLCResponseDecoder{}
	utils.Debug.Panic(decoder.InitFromHex(hexStr))
	assert.Equal(t, DateTimeStreamCode, decoder.GetIdentifier())
	dateTime, err := decoder.GetDateTime()
	utils.Debug.Panic(err)
	fmt.Printf("%s\n", dateTime.String())
}
