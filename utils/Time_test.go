package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeUtil_IsExpired(t *testing.T) {
	assert.True(t, Time.IsExpired(time.Now(), time.Time{}, time.Duration(10)*time.Second))
}
