package port

import (
	"fmt"
	"testing"
)

func TestObjectClassType_ToString(t *testing.T) {
	car := OctCar
	fmt.Println(car.ToString())
}
