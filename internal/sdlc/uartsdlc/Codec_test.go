package uartsdlc

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"rvpro3/radarvision.com/utils"
)

func TestDecode(t *testing.T) {
	cycle(t, "024218091a0c393add7d2203")
	cycle(t, "0213004B2F03")
	cycle(t, "024300000000000000000022EC03")
	cycle(t, "02117D230000000000000068DA03")
	cycle(t, "02117D230000000000000068DA03")
}

func decode(source string) {
	buffer, err := hex.DecodeString(source)
	utils.Debug.Panic(err)

	// Always safe, as rawData can only shrink
	res, err := Codec.Decode(buffer)
	utils.Debug.Panic(err)

	fmt.Println("Source:", source)
	fmt.Println("Result:", hex.EncodeToString(res))
}

func cycle(t *testing.T, expected string) {
	expected = strings.ToLower(expected)
	buffer, err := hex.DecodeString(expected)
	utils.Debug.Panic(err)

	sdlcReceive, err := Codec.Decode(buffer)
	utils.Debug.Panic(err)
	sdlcReceiveStr := strings.ToLower(hex.EncodeToString(sdlcReceive))

	sdlcSend, err := Codec.Encode(sdlcReceive[:len(sdlcReceive)-3])
	utils.Debug.Panic(err)
	sdlcSendStr := strings.ToLower(hex.EncodeToString(sdlcSend))

	assert.Equal(t, expected, sdlcSendStr)

	utils.Print.Ln("Source:", expected)
	utils.Print.Ln("Decode:", sdlcReceiveStr)
}

//func TestSendDetectRequest_EncodeTS2(t *testing.T) {
//	sdr := SendDetectRequest{}
//	reqData, err := sdr.EncodeTS2(4)
//	utils.Debug.Panic(err)
//
//	fmt.Println(hex.EncodeToString(reqData))
//}
