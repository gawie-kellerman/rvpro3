package golang

import (
	"fmt"
	"testing"
)

func TestArray(t *testing.T) {
	var arr [10]int

	fillArr(arr[:])
	printArr(arr[:])
}

func fillArr(ints []int) {
	for n := 1; n < 4; n++ {
		ints[n] = n
		//ints = append(ints, n)
	}
}

func printArr(ints []int) {
	fmt.Println(len(ints))
	fmt.Println(cap(ints))
	for i := range ints {
		fmt.Println(ints[i])
	}
}
