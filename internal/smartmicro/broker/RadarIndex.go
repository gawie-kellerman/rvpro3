package broker

import (
	"net"

	"rvpro3/radarvision.com/utils"
)

const Radar11IP = 3232238347

var GlobalRadarIPs [4]uint32

func RadarIndex(addr *net.UDPAddr) int {
	ip4 := utils.IP4Builder.FromIP(addr.IP, 0)
	ip4Int := ip4.ToU32()

	index := int(ip4Int) - Radar11IP

	// When index == 0 then radar 192.168.11.11

	if index >= 1 && index <= 4 {
		// Radar .12, .13, .14, .15
		// Index   0,   1,   2,  3
		index -= 1
	}

	if index >= 0 && index < len(GlobalRadarIPs) {
		GlobalRadarIPs[index] = ip4.ToU32()
		return index
	}
	return -1
}
