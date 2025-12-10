package sdlc

import (
	"encoding/binary"

	"rvpro3/radarvision.com/utils"
)

type SDLCRequestCode uint8

const StaticStatusRequestCode SDLCRequestCode = 0x10
const SendDetectDataCode SDLCRequestCode = 0x11
const ConfigBIURequestCode SDLCRequestCode = 0x12
const BIUDiagnosticRequestCode SDLCRequestCode = 0x13
const SDLCDiagnosticRequestCode SDLCRequestCode = 0x14
const DynamicStatusRequestCode SDLCRequestCode = 0x15
const SIUDiagnosticRequestCode SDLCRequestCode = 0x16

type SDLCRequestEncoder struct {
	buffer [64]byte
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
