package uartsdlc

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

type SDLCRequestCode uint8

type SDLCRequestEncoder struct {
	buffer [64]byte
}

func (s *SDLCRequestEncoder) GetIdentifier() SDLCIdentifier {
	return SDLCIdentifier(s.buffer[1])
}

func (s *SDLCRequestEncoder) BIUDiagnostics(reset byte) ([]byte, error) {
	fb := utils.FixedBuffer{Buffer: s.buffer[:]}
	fb.WriteU8(startMarker)
	fb.WriteU8(uint8(BIUDiagnosticRequestCode))
	fb.WriteU8(reset)
	return Codec.Encode(fb.AsWriteSlice())
}

func (s *SDLCRequestEncoder) DynamicStatus() ([]byte, error) {
	fb := utils.FixedBuffer{Buffer: s.buffer[:]}
	fb.WriteU8(startMarker)
	fb.WriteU8(uint8(DynamicStatusRequestCode))
	return Codec.Encode(fb.AsWriteSlice())
}

func (s *SDLCRequestEncoder) Diagnostics(reset byte) ([]byte, error) {
	fb := utils.FixedBuffer{Buffer: s.buffer[:]}
	fb.WriteU8(startMarker)
	fb.WriteU8(uint8(SDLCDiagnosticRequestCode))
	fb.WriteU8(reset)
	return Codec.Encode(fb.AsWriteSlice())
}

// TS2Detect is 8 bytes where
// byte 0..1 = BIU 1
// byte 2..3 = BIU 2
// byte 4..5 = BIU 3
// byte 6..7 = BIU 4
// This is a Little Endian presentation
func (s *SDLCRequestEncoder) TS2Detect(detects uint64) ([]byte, error) {
	fb := utils.FixedBuffer{Buffer: s.buffer[:]}
	fb.WriteU8(startMarker)
	fb.WriteU8(uint8(SendDetectDataCode))
	fb.WriteU64(detects, binary.LittleEndian)
	return Codec.Encode(fb.AsWriteSlice())
}

func (s *SDLCRequestEncoder) SIUDiagnostics(reset byte) ([]byte, error) {
	fb := utils.FixedBuffer{Buffer: s.buffer[:]}
	fb.WriteU8(startMarker)
	fb.WriteU8(uint8(SIUDiagnosticRequestCode))
	fb.WriteU8(reset)
	return Codec.Encode(fb.AsWriteSlice())
}

func (s *SDLCRequestEncoder) StaticStatus() ([]byte, error) {
	fb := utils.FixedBuffer{Buffer: s.buffer[:]}
	fb.WriteU8(startMarker)
	fb.WriteU8(uint8(StaticStatusRequestCode))
	return Codec.Encode(fb.AsWriteSlice())
}

func (s *SDLCRequestEncoder) ConfigBIU(flags byte) ([]byte, error) {
	fb := utils.FixedBuffer{Buffer: s.buffer[:]}
	fb.WriteU8(startMarker)
	fb.WriteU8(1) // DataLen
	fb.WriteU8(uint8(ConfigBIURequestCode))
	fb.WriteU8(flags)

	return Codec.Encode(fb.AsWriteSlice())
}
