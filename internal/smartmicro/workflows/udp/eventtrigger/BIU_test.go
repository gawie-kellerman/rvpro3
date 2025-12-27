package eventtrigger

import (
	"fmt"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"rvpro3/radarvision.com/utils/bit"
)

func TestBIU_ToLE(t *testing.T) {
	for n := 0; n <= 15; n++ {
		b := BIU(0).FromFlags(byte(n)) // 0xFFFF000000000000
		log.Info().
			Str("flags", fmt.Sprintf("%2d", n)).
			Str("bits", bit.AsString(b)).
			Str("le", bit.AsString(b.ToLE())).
			Msg("")
		assert.Equal(t, b.ToLE(), b.ToLEOld())
	}
}
