package port

type ObjectClassType uint8

const isNewObject = 1

const (
	OctUndefined ObjectClassType = iota
	OctPedestrian
	OctBicycle
	OctMotorbike
	OctCar
	OctReserved
	OctDelivery
	OctShortTruck
	OctLongTruck
)

func (o ObjectClassType) String() string {
	switch o {
	case OctUndefined:
		return "UNDEFINED"
	case OctPedestrian:
		return "PEDESTRIAN"
	case OctBicycle:
		return "BICYCLE"
	case OctMotorbike:
		return "MOTORBIKE"
	case OctCar:
		return "CAR"
	case OctReserved:
		return "RESERVED"
	case OctDelivery:
		return "DELIVERY TRUCK"
	case OctShortTruck:
		return "SHORT TRUCK"
	case OctLongTruck:
		return "LONG TRUCK"
	default:
		return "UNDEFINED"
	}
}
