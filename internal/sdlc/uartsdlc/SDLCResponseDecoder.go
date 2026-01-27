package uartsdlc

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"rvpro3/radarvision.com/utils"
)

type StaticStatusMode byte

func (s StaticStatusMode) IsATC() bool {
	return s&0x1 == 0x1
}

func (s StaticStatusMode) IsTS2() bool {
	return s&0x1 == 0x0
}

func (s StaticStatusMode) IsNormal() bool {
	return s&0x2 == 0x2
}

func (s StaticStatusMode) IsSafe() bool {
	return s&0x2 == 0x00
}

func (s StaticStatusMode) IsWdt() bool {
	return s&(1<<7) != 0
}

func (s StaticStatusMode) String() string {
	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("%08b, %d", int(s), int(s)))

	if s.IsATC() {
		str.WriteString(",ATC")
	} else {
		str.WriteString(",TS2")
	}

	if s.IsNormal() {
		str.WriteString(",Normal")
	} else {
		str.WriteString(",Safe")
	}

	if s.IsWdt() {
		str.WriteString(",WDT")
	}
	return str.String()
}

type SDLCResponseDecoder struct {
	rawData [256]byte
	slice   []byte
}

type BIUDiagnostics struct {
	MMULoadSwitchCounter     byte
	DateTimeBroadcastCounter byte
	CallDataRequestCounter   [4]byte
	ResetDiagnosticCounter   [4]byte
	ServiceRequestCounter    byte
	ReservedCounter          byte
}

func (d BIUDiagnostics) PrintDetail() {
	utils.Print.Detail("BIU Diagnostics Response", "\n")
	utils.Print.Indent(2)

	utils.Print.Detail("MMU MergeFromFile Switch", "%d\n", d.MMULoadSwitchCounter)
	utils.Print.Detail("DateTime Broadcast", "%d\n", d.DateTimeBroadcastCounter)

	for counter := range d.CallDataRequestCounter {
		utils.Print.Detail("Call data Request ", "%d\n", counter)
	}

	for counter := range d.ResetDiagnosticCounter {
		utils.Print.Detail("Reset Diagnostic ", "%d\n", counter)
	}

	utils.Print.Detail("Service Request", "%d\n", d.ServiceRequestCounter)
	utils.Print.Detail("Reserved", "%d\n", d.ReservedCounter)

	utils.Print.Indent(-2)
}

type SIUDiagnostics struct {
	StatusCounter            byte
	MillisecondCounter       byte
	InputConfigCounter       byte
	PollRawCounter           byte
	PollFilteredCounter      byte
	TransitionBufferCounter  byte
	ModuleIDCounter          byte
	TimeDateCounter          byte
	ModuleDescriptionCounter byte
}

type DynamicStatus struct {
	SinceLastSDLCComms int
	RequestedCount     byte
	UptimeInDays       byte
	UptimeIn6Mins      int
	SdlcFailCount      int
	UartFailCount      int
	IsFailSafeMapped   bool
}

func (s DynamicStatus) PrintDetail() {
	utils.Print.Detail("Dynamic Status Response", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("SDLC Fail Metric", "%d\n", s.SdlcFailCount)
	utils.Print.Detail("Request Metric", "%d\n", s.RequestedCount)
	utils.Print.Detail("UART Fail Metric", "%d\n", s.UartFailCount)
	utils.Print.Detail("Since Last COMMs", "%d\n", s.SinceLastSDLCComms)
	utils.Print.Detail("Uptime (Days)", "%d\n", s.UptimeInDays)
	utils.Print.Detail("Uptime (mins)", "%d\n", s.UptimeIn6Mins)
	utils.Print.Indent(-2)
}

type SDLCDiagnostics struct {
	ShortFrameError int
	ControlError    int
	CRCError        int
	IdleError       int
	FramingError    int
	LongFrameError  int
}

func (d SDLCDiagnostics) PrintDetail() {
	utils.Print.Detail("SDLC Diagnostics Response", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("Control Error", "%d\n", d.ControlError)
	utils.Print.Detail("CRC Error", "%d\n", d.CRCError)
	utils.Print.Detail("IdleError", "%d\n", d.IdleError)
	utils.Print.Detail("Framing Error", "%d\n", d.FramingError)
	utils.Print.Detail("Long-Frame Error", "%d\n", d.LongFrameError)
	utils.Print.Indent(-2)
}

type StaticStatus struct {
	BIU             BIUFlags
	MajorVersion    byte
	MinorVersion    byte
	Serial          uint64
	ProtocolVersion byte
	Mode            StaticStatusMode
	IsModeMapped    bool
}

func (s *StaticStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"BIU":             s.BIU,
		"BIUMaskHO":       s.BIU.String(),
		"MajorVersion":    s.MajorVersion,
		"MinorVersion":    s.MinorVersion,
		"Serial":          fmt.Sprintf("%016x", s.Serial),
		"ProtocolVersion": s.ProtocolVersion,
		"Mode":            s.Mode.String(),
	})
}

func (s *StaticStatus) PrintDetail() {
	utils.Print.Detail("SDLC Static Status", "\n")
	utils.Print.Indent(2)
	utils.Print.Detail("SDLC Version", "%d.%d\n", int(s.MajorVersion), int(s.MinorVersion))
	utils.Print.Detail("Protocol Version", "%d\n", s.ProtocolVersion)
	utils.Print.Detail("Mode", "%d, %s\n", int(s.Mode), s.Mode.String())
	utils.Print.Detail("BIU", "%08b\n", int(s.BIU))
	utils.Print.Detail("Serial", "%16x\n", s.Serial)
	utils.Print.Indent(-2)
}

type CMUFrame struct {
	Mode     StaticStatusMode
	Green    uint32
	Yellow   uint32
	Red      uint32
	BitCount int
}

func (c *CMUFrame) String() string {
	if c.Mode.IsTS2() {
		return fmt.Sprintf("Mode: %s, Green: %04x, Yellow: %04x, Red: %04x", c.Mode, c.Green, c.Yellow, c.Red)
	}
	return fmt.Sprintf("Mode: %s, Green: %08x, Yellow: %08x, Red: %08x", c.Mode, c.Green, c.Yellow, c.Red)

}

func (s *SDLCResponseDecoder) Init(buffer []byte) (err error) {
	s.slice, err = Codec.DecodeInto(buffer, s.rawData[:])
	return err
}

func (s *SDLCResponseDecoder) GetIdentifier() SDLCIdentifier {
	return SDLCIdentifier(s.slice[1])
}

func (s *SDLCResponseDecoder) GetBIUDiagnostics() (BIUDiagnostics, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}

	res := BIUDiagnostics{}
	res.MMULoadSwitchCounter = fb.ReadU8()
	res.DateTimeBroadcastCounter = fb.ReadU8()
	res.CallDataRequestCounter[0] = fb.ReadU8()
	res.CallDataRequestCounter[1] = fb.ReadU8()
	res.CallDataRequestCounter[2] = fb.ReadU8()
	res.CallDataRequestCounter[3] = fb.ReadU8()
	res.ResetDiagnosticCounter[0] = fb.ReadU8()
	res.ResetDiagnosticCounter[1] = fb.ReadU8()
	res.ResetDiagnosticCounter[2] = fb.ReadU8()
	res.ResetDiagnosticCounter[3] = fb.ReadU8()
	res.ServiceRequestCounter = fb.ReadU8()
	res.ReservedCounter = fb.ReadU8()

	return res, fb.Err
}

func (s *SDLCResponseDecoder) GetCMUFrame() (CMUFrame, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}

	res := CMUFrame{}
	res.Mode = StaticStatusMode(fb.ReadU8())

	if res.Mode.IsTS2() {
		res.Green = uint32(fb.ReadU16(binary.BigEndian))
		res.Yellow = uint32(fb.ReadU16(binary.BigEndian))
		res.Red = uint32(fb.ReadU16(binary.BigEndian))
		res.BitCount = 16
	} else {
		res.Green = fb.ReadU32(binary.BigEndian)
		res.Yellow = fb.ReadU32(binary.BigEndian)
		res.Red = fb.ReadU32(binary.BigEndian)
		res.BitCount = 32
	}

	return res, fb.Err
}

func (s *SDLCResponseDecoder) GetAcknowledge() (byte, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}
	return fb.ReadU8(), fb.Err
}

func (s *SDLCResponseDecoder) GetDateTime() (time.Time, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}

	return time.Date(
		2000+int(fb.ReadU8()),
		time.Month(fb.ReadU8()),
		int(fb.ReadU8()),
		int(fb.ReadU8()),
		int(fb.ReadU8()),
		int(fb.ReadU8()),
		0,
		time.Local,
	), fb.Err
}

func (s *SDLCResponseDecoder) GetDynamicStatus() (DynamicStatus, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}

	res := DynamicStatus{}
	dataLen := Codec.GetDataLen(s.slice)

	// TODO: Check the spec
	res.SinceLastSDLCComms = int(fb.ReadU8()) * 10
	res.RequestedCount = fb.ReadU8()
	res.UptimeInDays = fb.ReadU8()
	res.UptimeIn6Mins = int(fb.ReadU8()) * 6

	switch dataLen {
	case 5:
		res.SdlcFailCount = int(fb.ReadU8())
		res.IsFailSafeMapped = true

	case 6:
		res.SdlcFailCount = int(fb.ReadU8())
		res.UartFailCount = int(fb.ReadU8())

	default:
		// Do nothing
	}

	return res, fb.Err
}

func (s *SDLCResponseDecoder) GetSDLCDiagnostics() (SDLCDiagnostics, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}
	return SDLCDiagnostics{
		ShortFrameError: calcCount(fb.ReadU8()),
		ControlError:    calcCount(fb.ReadU8()),
		CRCError:        calcCount(fb.ReadU8()),
		IdleError:       calcCount(fb.ReadU8()),
		FramingError:    calcCount(fb.ReadU8()),
		LongFrameError:  calcCount(fb.ReadU8()),
	}, fb.Err
}

func calcCount(count byte) int {
	if count < 129 {
		return int(count)
	}
	return (int(count) - 127) * 128
}

func (s *SDLCResponseDecoder) GetSIUDiagnostics() (SIUDiagnostics, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}
	return SIUDiagnostics{
		StatusCounter:            fb.ReadU8(),
		MillisecondCounter:       fb.ReadU8(),
		InputConfigCounter:       fb.ReadU8(),
		PollRawCounter:           fb.ReadU8(),
		PollFilteredCounter:      fb.ReadU8(),
		TransitionBufferCounter:  fb.ReadU8(),
		ModuleIDCounter:          fb.ReadU8(),
		TimeDateCounter:          fb.ReadU8(),
		ModuleDescriptionCounter: fb.ReadU8(),
	}, fb.Err
}

func (s *SDLCResponseDecoder) GetStaticStatus() (StaticStatus, error) {
	fb := utils.FixedBuffer{Buffer: s.slice, WritePos: len(s.slice), ReadPos: 2}

	return StaticStatus{
		BIU:             BIUFlags(fb.ReadU8()),
		MajorVersion:    fb.ReadU8(),
		MinorVersion:    fb.ReadU8(),
		Serial:          fb.ReadU64(binary.LittleEndian),
		ProtocolVersion: fb.ReadU8(),
		Mode:            StaticStatusMode(fb.ReadU8()),
		IsModeMapped:    Codec.GetDataLen(s.slice) == 14,
	}, fb.Err
}

func (s *SDLCResponseDecoder) InitFromHex(hexStr string) error {
	hexStr = strings.ToUpper(hexStr)

	if strings.HasPrefix(hexStr, "0X") {
		hexStr = hexStr[2:]
	}
	buffer, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}
	return s.Init(buffer)
}
