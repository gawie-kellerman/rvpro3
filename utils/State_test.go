package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState_Get(t *testing.T) {
	s := State{}
	s.Init()

	assert.Nil(t, s.Get("NotExistKey"))
}
