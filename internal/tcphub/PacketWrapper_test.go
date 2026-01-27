package tcphub

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"rvpro3/radarvision.com/utils"
)

func TestPacketWrapper_Properties(t *testing.T) {
	var buffer [300]byte

	pw := PacketWrapper{}
	pw.Init(buffer[:], 1, utils.IP4Builder.FromU32(100, 55555))
	pw.SetPacketType(PtUdpForward)
	pw.SetData([]byte("Hello World"))

	fmt.Println(pw.GetPacketSize())
	pw.Buffer = buffer[:pw.GetPacketSize()]
	assert.Truef(t, pw.IsParseableLength(), "IsParseableLength should be true")
	assert.True(t, pw.IsComplete())
	assert.Equal(t, uint16(2), pw.GetVersion())
	assert.Equal(t, startDelimiter, pw.GetDelimiter())
	assert.Equal(t, uint32(1), pw.GetSequence())
	assert.Equal(t, PtUdpForward, pw.GetPacketType())
	assert.Equal(t, uint32(100), pw.GetTargetIP())
	assert.Equal(t, uint16(55555), pw.GetTargetPort())
	assert.Equal(t, uint16(len("Hello World")), pw.GetDataSize())
	assert.Equal(t, "Hello World", string(pw.GetData()))
}
