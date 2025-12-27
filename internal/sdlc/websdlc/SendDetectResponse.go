package websdlc

import (
	"fmt"
	"strconv"
	"strings"
)

type SendDetectResponse struct {
	Ok         int64
	Red        uint32
	Yellow     uint32
	Green      uint32
	Call       string
	SdlcMillis int
}

// Unmarshal / sample
// /OK:4<br>
// /RED:00000000<br>
// /YEL:00000000<br>
// /GRN:00000000<br>
// /CALL:2000000000000000000000000000000000000000000000000000000000000000<br>
// /SDLC:2000
// /Notes:
// /1. Lines terminated with 0x0a0d (Windows)
// /2. SDLC Line does not have a terminator
func (r *SendDetectResponse) Unmarshal(body string) (err error) {
	var i64 int64

	parts := strings.Split(body, "\r\n")

	for _, part := range parts {
		kvPair := strings.Split(part, ":")

		if len(kvPair) == 2 {
			kvPair[1] = strings.TrimRight(kvPair[1], "\r\n")
			kvPair[1] = strings.TrimRight(kvPair[1], "<br>")

			fmt.Println(kvPair[0], kvPair[1])
			switch kvPair[0] {
			case "OK":
				r.Ok, err = strconv.ParseInt(kvPair[1], 10, 32)

			case "RED":
				i64, err = strconv.ParseInt(kvPair[1], 16, 32)
				r.Red = uint32(i64)

			case "YEL":
				i64, err = strconv.ParseInt(kvPair[1], 16, 32)
				r.Yellow = uint32(i64)

			case "GRN":
				i64, err = strconv.ParseInt(kvPair[1], 16, 32)
				r.Green = uint32(i64)

			case "CALL":
				r.Call = kvPair[1]

			case "SDLC":
				i64, err = strconv.ParseInt(kvPair[1], 10, 32)
				r.SdlcMillis = int(i64)
			}

			if err != nil {
				return err
			}
		}
	}

	return err
}
