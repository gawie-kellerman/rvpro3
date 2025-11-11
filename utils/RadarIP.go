package utils

var Radar11IP IP4

func init() {
	Radar11IP = IP4Builder.FromOctets(192, 168, 11, 11, 55555)
}

func RadarIndexOf(ipAddr uint32) int {
	distance := Radar11IP.DistanceTo(ipAddr)

	if distance < 0 || distance > 5 {
		return -1
	}

	if distance >= 1 {
		return distance - 1
	} else {
		return distance
	}
}

func RadarIPOf(index int) IP4 {
	return Radar11IP.WithHost8(12 + index)
}
