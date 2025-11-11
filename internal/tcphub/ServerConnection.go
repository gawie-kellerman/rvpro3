package tcphub

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/rs/zerolog/log"

	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

type ServerConnection struct {
	service.MixinDataService
	writeChannel chan Packet
	Connection   net.TCPConn
	Stats        ServerConnectionStats
	Buffer       PacketBuffer
	writeCache   [1500]byte
}

func (client *ServerConnection) start(conn net.TCPConn) {
	client.Terminated = false
	client.Terminating = false
	client.Connection = conn
	client.Buffer.Init(8 * utils.Kilobyte)
	client.Stats.Start(client.Connection.RemoteAddr())
	client.writeChannel = make(chan Packet, 10)
	go client.executeRead()
	go client.executeWrite()
}

func (client *ServerConnection) executeRead() {
	var readBuffer [8 * utils.Kilobyte]byte

	reader := bufio.NewReader(&client.Connection)

	for client.Terminating = false; !client.Terminating; {
		bytesRead, err := reader.Read(readBuffer[:])

		if err != nil {
			if err != io.EOF {
				log.Err(err).
					Str("Addr", client.Connection.RemoteAddr().String()).
					Msg("TCP Server ServerConnection Read Error")
				client.Terminating = true
				client.Stats.ReadStats.ErrCount += uint64(bytesRead)
			}
		} else {
			if bytesRead > 0 {
				client.Stats.ReadStats.ByteCount += uint64(bytesRead)
				client.Stats.ReadStats.CycleCount++
				client.handle(readBuffer[:bytesRead])
			}
		}
	}

	client.writeChannel <- Packet{}
}

func (client *ServerConnection) handle(buffer []byte) {
	client.Buffer.PushBytes(buffer)

	packet := Packet{}

	for ; ; client.Buffer.Pop(&packet) {
		fmt.Println("Packet sent to radar")
		//SendToRadar(packet)
	}
}

func (client *ServerConnection) closeConnection() {
	client.Stats.Stop()
	if client.OnCloseConnection != nil {
		client.OnCloseConnection(client)
	}
	_ = client.Connection.Close()
}

func (client *ServerConnection) executeWrite() {
	for packet := range client.writeChannel {
		if packet.Size == 0 {
			break
		}
		client.writePacket(packet)
	}
	client.closeConnection()
	client.Terminated = true
	client.Stats.IsOpen = false
	close(client.writeChannel)
}

func (client *ServerConnection) Write(packet Packet) bool {
	if client.Terminating {
		return false
	}

	if len(client.writeChannel) < cap(client.writeChannel)/2 {
		client.writeChannel <- packet
		return true
	} else {
		if packet.Size == 0 || packet.Type == PtUdpInstruction {
			client.writeChannel <- packet
			return true
		}
	}
	return false
}

func (client *ServerConnection) writePacket(packet Packet) {
	if client.Terminating {
		return
	}

	slice, err := packet.SaveToBytes(client.writeCache[:])

	if err != nil {
		client.Stats.WriteStats.OverflowCount++
		log.Err(err).
			Str("Addr", client.Connection.RemoteAddr().String()).
			Msg("Packet slice too big")
		return
	}

	bytesWritten, err := client.Connection.Write(slice)
	if err != nil {
		client.Stats.WriteStats.WriteErrCount++
		log.Err(err).
			Msg("TCP Server Write Error")
		client.Terminating = true
	} else {
		if bytesWritten != len(slice) {
			log.Err(err).
				Msg("TCP Server Write Size difference")
		} else {
			client.Stats.WriteStats.WriteOKCount++
			client.Stats.WriteStats.WriteOKSize += uint64(bytesWritten)
		}
	}
}
