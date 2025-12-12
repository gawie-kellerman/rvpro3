package sdlc

import (
	"encoding/hex"
	"fmt"
	"math"
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

func TestDunno(t *testing.T) {
	decoder := SDLCResponseDecoder{}
	utils.Debug.Panic(decoder.InitFromHex("02110000000000000000A0AF03"))
	fmt.Println(decoder.GetIdentifier().String())
}

func TestSDLCRequestEncoder_TS2Detect(t *testing.T) {
	encode := func(value uint64, expect string) {
		encoder := SDLCRequestEncoder{}
		detect, err := encoder.TS2Detect(value)
		utils.Debug.Panic(err)
		assert.Equal(t, expect, hex.EncodeToString(detect))
	}

	encode(7, "0211070000000000000067b703")
	encode(127000, "021118f0010000000000e9f903")
	encode(math.MaxUint64, "0211ffffffffffffffff064e03")
	encode(0, "02110000000000000000a0af03")
}
