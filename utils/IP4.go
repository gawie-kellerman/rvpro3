package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
)

var IP4Builder ip4Builder

type ip4Builder struct {
}

func (ip4Builder) FromOctets(o1 byte, o2 byte, o3 byte, o4 byte, port int) IP4 {
	return IP4{
		bytes: [4]byte{o1, o2, o3, o4},
		Port:  port,
	}
}

func (ip4Builder) FromString(addr string) IP4 {
	ipPart := getIPPart(addr)
	port := getPortFromString(addr)

	ip := net.ParseIP(ipPart)
	return IP4Builder.FromIP(ip, port)
}

func (ip4Builder) FromIP(addr net.IP, port int) IP4 {
	result := IP4{}

	if len(addr) == 16 {
		copy(result.bytes[:], addr[12:16])
	} else {
		copy(result.bytes[:], addr[0:4])
	}
	result.Port = port
	return result
}

func (ip4Builder) FromBytes(bytes []byte, port int) IP4 {
	result := IP4{}
	result.Port = port
	copy(result.bytes[:], bytes)
	return result
}

func (ip4Builder) FromU32(ip uint32, port int) IP4 {
	result := IP4{}
	result.Port = port
	binary.BigEndian.PutUint32(result.bytes[0:4], ip)
	return result
}

func (b ip4Builder) FromAddr(addr net.Addr) IP4 {
	if udpAddr, ok := addr.(*net.UDPAddr); ok {
		return IP4Builder.FromIP(udpAddr.IP, udpAddr.Port)
	}

	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		return IP4Builder.FromIP(tcpAddr.IP, tcpAddr.Port)
	}

	result := IP4{}
	return result
}

func (b ip4Builder) FromPort(port int) IP4 {
	res := IP4{}
	res.Port = port
	return res
}

type IP4 struct {
	bytes [4]byte
	Port  int
}

func (i IP4) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"RadarIP": i.String(),
		"Port":    i.Port,
	})
}

func (i IP4) String() string {
	res := strings.Builder{}
	res.Grow(20)

	res.WriteString(fmt.Sprintf(
		"%d.%d.%d.%d:%d",
		i.bytes[0],
		i.bytes[1],
		i.bytes[2],
		i.bytes[3],
		i.Port,
	))
	return res.String()
}

func (i IP4) ToUDPAddr() net.UDPAddr {
	res := net.UDPAddr{
		IP:   net.IPv4(i.bytes[0], i.bytes[1], i.bytes[2], i.bytes[3]),
		Port: i.Port,
	}

	return res
}

func (i IP4) ToU32() uint32 {
	return binary.BigEndian.Uint32(i.bytes[:])
}

func (i IP4) WithUDPPort(port int) net.UDPAddr {
	res := net.UDPAddr{
		IP:   net.IPv4(i.bytes[0], i.bytes[1], i.bytes[2], i.bytes[3]),
		Port: port,
	}

	return res
}

func (i IP4) WithPort(port int) IP4 {
	return IP4{
		bytes: i.bytes,
		Port:  port,
	}
}

func (i IP4) RelativeTo(baseIP uint32) int {
	return int(i.ToU32()) - int(baseIP)
}

func (i IP4) ToTCPAddr() net.TCPAddr {
	res := net.TCPAddr{
		IP:   net.IPv4(i.bytes[0], i.bytes[1], i.bytes[2], i.bytes[3]),
		Port: i.Port,
	}

	return res
}

func (i IP4) DistanceTo(ip4 uint32) int {
	return int(ip4) - int(i.ToU32())
}

func (i IP4) WithHost8(host int) IP4 {
	i.bytes[3] = byte(host)
	return i
}

func (i IP4) ToIPString() string {
	res := strings.Builder{}
	res.Grow(20)

	res.WriteString(fmt.Sprintf(
		"%d.%d.%d.%d",
		i.bytes[0],
		i.bytes[1],
		i.bytes[2],
		i.bytes[3],
	))
	return res.String()

}

func (i IP4) WriteToFixedBuffer(writer *FixedBuffer) {
	writer.WriteBytes(i.bytes[:])
	writer.WriteU16(uint16(i.Port), binary.BigEndian)
}

func (i IP4) Equals(other IP4) bool {
	return other.ToU32() == i.ToU32() && other.Port == i.Port
}

func (i IP4) IsEqualIP(other IP4) bool {
	return other.ToU32() == i.ToU32()
}

func getPortFromString(addr string) int {
	port, err := strconv.Atoi(getPortPart(addr))
	if err != nil {
		return 0
	}
	return port
}

func getIPPart(addr string) string {
	index := strings.Index(addr, ":")

	if index == -1 {
		return addr
	}
	return addr[:index]
}

func getPortPart(addr string) string {
	index := strings.Index(addr, ":")

	if index == -1 {
		return addr
	}
	return addr[index+1:]
}
