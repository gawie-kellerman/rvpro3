package port

import (
	"testing"

	"rvpro3/radarvision.com/utils"
)

func TestObjectClassType_ToString(t *testing.T) {
	car := OctCar
	utils.Debug.Println(car)
}
