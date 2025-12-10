package port

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlags_SetFlag(t *testing.T) {
	value := FlagsType(0)
	value.Set(FlMessageCount)
	assert.Equal(t, 1, value)
}

func TestFlags_Compound(t *testing.T) {
	expected := FlMessageCount | FlTimestamp
	value := FlagsType(0)
	value.Set(FlMessageCount)
	value.Set(FlTimestamp)
	assert.Equal(t, expected, value)
	fmt.Println(value.String())
}
