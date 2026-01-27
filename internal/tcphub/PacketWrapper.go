package tcphub

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"rvpro3/radarvision.com/utils"
)

const delimiterOffset = 0
const versionOffset = delimiterOffset + 4
const sequenceOffset = versionOffset + 2
const timeStampOffset = sequenceOffset + 4
const dataSizeOffset = timeStampOffset + 8
const packetTypeOffset = dataSizeOffset + 2
const targetIP4Offset = packetTypeOffset + 1
const targetPortOffset = targetIP4Offset + 4
const sourceIP4Offset = targetPortOffset + 2
const sourcePortOffset = sourceIP4Offset + 4
const dataOffset = sourcePortOffset + 2

type PacketWrapper struct {
	Buffer []byte
}

func (p *PacketWrapper) Init(buffer []byte, sequenceNo uint32, targetIP utils.IP4, sourceIP utils.IP4) {
	p.Buffer = buffer

	p.SetDelimiter()
	p.SetVersion(2)
	p.SetSequence(sequenceNo)
	p.SetTimeStamp(time.Now().UnixMilli())
	p.SetTargetIP(targetIP.ToU32())
	p.SetTargetPort(uint16(targetIP.Port))
	p.SetSourceIP(sourceIP.ToU32())
	p.SetSourcePort(uint16(sourceIP.Port))
}

func (p *PacketWrapper) IsParseableLength() bool {
	return len(p.Buffer) >= dataSizeOffset
}

func (p *PacketWrapper) IsValidStart() bool {
	return p.GetDelimiter() == startDelimiter
}

func (p *PacketWrapper) IsComplete() bool {
	if p.IsParseableLength() {
		if len(p.Buffer) < p.GetPacketSize() {
			return false
		}
		return true
	}
	return false
}

func (p *PacketWrapper) GetDelimiter() uint32 {
	return binary.LittleEndian.Uint32(p.Buffer[delimiterOffset : delimiterOffset+4])
}

func (p *PacketWrapper) SetDelimiter() {
	binary.LittleEndian.PutUint32(p.Buffer[delimiterOffset:], startDelimiter)
}

func (p *PacketWrapper) GetVersion() uint16 {
	return binary.LittleEndian.Uint16(p.Buffer[versionOffset : versionOffset+2])
}

func (p *PacketWrapper) GetSequence() uint32 {
	return binary.LittleEndian.Uint32(p.Buffer[sequenceOffset : sequenceOffset+4])
}

func (p *PacketWrapper) SetSequence(value uint32) {
	binary.LittleEndian.PutUint32(p.Buffer[sequenceOffset:], value)
}

func (p *PacketWrapper) SetVersion(version uint16) {
	binary.LittleEndian.PutUint16(p.Buffer[versionOffset:], version)
}

func (p *PacketWrapper) GetTimeStamp() int64 {
	return int64(binary.LittleEndian.Uint64(p.Buffer[timeStampOffset : timeStampOffset+8]))
}

func (p *PacketWrapper) SetTimeStamp(timeStamp int64) {
	binary.LittleEndian.PutUint64(p.Buffer[timeStampOffset:], uint64(timeStamp))
}

func (p *PacketWrapper) GetDataSize() uint16 {
	return binary.LittleEndian.Uint16(p.Buffer[dataSizeOffset : dataSizeOffset+2])
}

func (p *PacketWrapper) SetDataSize(dataSize uint16) {
	binary.LittleEndian.PutUint16(p.Buffer[dataSizeOffset:], dataSize)
}

func (p *PacketWrapper) GetPacketType() PacketType {
	return PacketType(p.Buffer[packetTypeOffset])
}

func (p *PacketWrapper) SetPacketType(packetType PacketType) {
	p.Buffer[packetTypeOffset] = byte(packetType)
}

func (p *PacketWrapper) GetTargetIP() uint32 {
	return binary.LittleEndian.Uint32(p.Buffer[targetIP4Offset : targetIP4Offset+4])
}

func (p *PacketWrapper) GetSourceIP() uint32 {
	return binary.LittleEndian.Uint32(p.Buffer[sourceIP4Offset : sourceIP4Offset+4])
}

func (p *PacketWrapper) SetSourceIP(sourceIP uint32) {
	binary.LittleEndian.PutUint32(p.Buffer[sourceIP4Offset:], sourceIP)
}

func (p *PacketWrapper) GetSourcePort() uint16 {
	return binary.LittleEndian.Uint16(p.Buffer[sourcePortOffset : sourcePortOffset+2])
}

func (p *PacketWrapper) SetSourcePort(sourcePort uint16) {
	binary.LittleEndian.PutUint16(p.Buffer[sourcePortOffset:], sourcePort)
}

func (p *PacketWrapper) SetTargetIP(targetIP uint32) {
	binary.LittleEndian.PutUint32(p.Buffer[targetIP4Offset:], targetIP)
}

func (p *PacketWrapper) GetTargetPort() uint16 {
	return binary.LittleEndian.Uint16(p.Buffer[targetPortOffset : targetPortOffset+2])
}

func (p *PacketWrapper) SetTargetPort(targetPort uint16) {
	binary.LittleEndian.PutUint16(p.Buffer[targetPortOffset:], targetPort)
}

func (p *PacketWrapper) GetPacket() []byte {
	return p.Buffer[:p.GetPacketSize()]
}

func (p *PacketWrapper) GetData() []byte {
	return p.Buffer[dataOffset : dataOffset+int(p.GetDataSize())]
}

func (p *PacketWrapper) SetData(data []byte) {
	p.SetDataSize(uint16(len(data)))
	copy(p.Buffer[dataOffset:], data)
}

func (p *PacketWrapper) IsDataFit(length int) bool {
	return len(p.Buffer) > dataOffset+length
}

func (p *PacketWrapper) GetPacketSize() int {
	return p.GetHeaderSize() + int(p.GetDataSize())
}

func (p *PacketWrapper) GetHeaderSize() int {
	return dataOffset + 1
}

func (p *PacketWrapper) GetTargetUDPAddr() net.UDPAddr {
	return p.GetTargetIP4().ToUDPAddr()
}

func (p *PacketWrapper) GetTargetIP4() utils.IP4 {
	return utils.IP4Builder.FromU32(p.GetTargetIP(), int(p.GetTargetPort()))
}

func (p *PacketWrapper) GetSourceIP4() utils.IP4 {
	return utils.IP4Builder.FromU32(p.GetSourceIP(), int(p.GetSourcePort()))
}

func (p *PacketWrapper) Dump(source string) {
	fmt.Printf(
		"%s => target: %s, source: %s, len: %d, data: %s\n",
		source,
		p.GetTargetIP4(),
		p.GetSourceIP4(),
		p.GetDataSize(),
		hex.EncodeToString(p.GetData()),
	)
}
