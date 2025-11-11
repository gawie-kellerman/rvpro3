package tcphub

import (
	"bytes"
	"encoding/binary"
	"net"
	"time"

	"rvpro3/radarvision.com/utils"
)

const startDelimiter uint32 = (62 << 24) | (62 << 16) | (62 << 8) | 123

// headerSize is 23
// 0-3 start delimiter
// 4-5 version
// 6-13 date
// 14-15 size
// 16 packet type
// 17-20 ip4 address
// 21-22 ip4 port
const headerSize = 23

type PacketType uint8

const (
	PtUnknown PacketType = iota
	PtUdpForward
	PtRadarMulticast
	PtUdpInstruction
	PtStats
	PtServerClosesConnection
)

type Packet struct {
	Delimiter  uint32
	Version    uint16
	Date       int64
	Size       uint16
	Type       PacketType
	TargetIP4  uint32
	TargetPort uint16
	Data       []byte
}

func (p *Packet) GetDataSize() int {
	return int(p.Size - headerSize)
}

func NewPacket(data []byte) Packet {
	res := Packet{
		Delimiter:  startDelimiter,
		Version:    1,
		Date:       utils.Time.ToLocalMillis(time.Now()),
		Type:       PtUdpForward,
		Size:       uint16(len(data) + headerSize),
		TargetIP4:  0,
		TargetPort: 55555,
		Data:       data,
	}
	return res
}

func (s *packetBuilder) Serialize(now time.Time, packetType PacketType, ip4 utils.IP4, payload []byte, target []byte) (int, error) {
	size := len(payload) + headerSize
	writer := utils.NewFixedBuffer(target, 0, 0)

	writer.WriteU32(startDelimiter, binary.LittleEndian)
	writer.WriteU16(1, binary.LittleEndian) //version
	writer.WriteI64(utils.Time.ToLocalMillis(now), binary.LittleEndian)
	writer.WriteU16(uint16(size), binary.LittleEndian)
	writer.WriteU8(uint8(packetType))
	writer.WriteU32(ip4.ToU32(), binary.LittleEndian)
	writer.WriteU16(uint16(ip4.Port), binary.LittleEndian)
	writer.WriteBytes(payload)

	return writer.WritePos, writer.Err
}

type packetBuilder struct {
}

var PacketBuilder packetBuilder

func (s *packetBuilder) Deserialize(target *Packet, source []byte) (int, error) {
	reader := utils.NewFixedBuffer(source, 0, len(source))

	target.Delimiter = reader.ReadU32(binary.LittleEndian)
	target.Version = reader.ReadU16(binary.LittleEndian)
	target.Date = int64(reader.ReadU64(binary.LittleEndian))
	target.Size = reader.ReadU16(binary.LittleEndian)
	target.Type = PacketType(reader.ReadU8())
	target.TargetIP4 = reader.ReadU32(binary.LittleEndian)
	target.TargetPort = reader.ReadU16(binary.LittleEndian)
	target.Data = reader.ReadBytes(target.GetDataSize())

	return reader.ReadPos, reader.Err
}

func (p *Packet) IsValid() bool {
	return p.Version == 1 && p.Delimiter == startDelimiter && len(p.Data) == p.GetDataSize()
}

func getPacketDelimiter(source []byte) uint32 {
	return binary.LittleEndian.Uint32(source[0:4])
}

func getPacketSize(source []byte) uint16 {
	return binary.LittleEndian.Uint16(source[14:16])
}

func (p *Packet) Write(writer *utils.FixedBuffer) {
	writer.WriteU32(startDelimiter, binary.LittleEndian)
	writer.WriteU16(1, binary.LittleEndian)
	writer.WriteI64(p.Date, binary.LittleEndian)
	writer.WriteU16(p.Size, binary.LittleEndian)
	writer.WriteU8(uint8(p.Type))
	writer.WriteU32(p.TargetIP4, binary.LittleEndian)
	writer.WriteU16(p.TargetPort, binary.LittleEndian)
	writer.WriteBytes(p.Data)
}

func (p *Packet) Equals(other *Packet) bool {
	return p.Date == other.Date &&
		p.Size == other.Size &&
		p.Version == other.Version &&
		p.Delimiter == other.Delimiter &&
		p.Type == other.Type &&
		p.TargetPort == other.TargetPort &&
		p.TargetIP4 == other.TargetIP4 &&
		bytes.Equal(p.Data, other.Data)
}

func (p *Packet) SaveToBytes(buffer []byte) ([]byte, error) {
	writer := utils.NewFixedBuffer(buffer, 0, 0)
	p.Write(&writer)
	return buffer[:writer.WritePos], writer.Err
}

func (p *Packet) GetTargetIP() utils.IP4 {
	return utils.IP4Builder.FromU32(p.TargetIP4, int(p.TargetPort))
}

func (p *Packet) SetTarget(addr *net.UDPAddr) {
	ip4 := utils.IP4Builder.FromAddr(addr)
	p.TargetIP4 = ip4.ToU32()
	p.TargetPort = uint16(ip4.Port)
}
