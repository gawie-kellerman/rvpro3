package interfaces

import "rvpro3/radarvision.com/utils"

type IUDPWriter interface {
	WriteData(ip4 utils.IP4, data []byte) error
}
