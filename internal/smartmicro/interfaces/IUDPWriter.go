package interfaces

type IUDPPacketWriter interface {
	WritePacket(packet []byte) error
}
