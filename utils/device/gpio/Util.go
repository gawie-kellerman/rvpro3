package gpio

import "strconv"

type util struct{}

var Util util

func (util) GetChipNo(portNo int) int {
	return portNo / 32
}

func (util) GetChipName(chipNo int) string {
	return "gpiochip" + strconv.Itoa(chipNo)
}

func (util) GetOffset(portNo int) int {
	return portNo % 32
}

func (util) GetPortNo(chipNo int, offset int) int {
	return chipNo*32 + offset
}
