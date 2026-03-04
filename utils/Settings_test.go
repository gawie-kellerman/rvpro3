package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const xKey = "radar.udp.verbose.trigger"
const xIndex = "[192]"

//const xFull = "radar.udp.verbose.trigger-[192]"

func TestSettings_Indexed(t *testing.T) {
	sets := &Settings{}
	sets.Init()

	sets.Basic.Set("sample", "value")
	sets.Indexed.SetDefault(xKey, "false")
	assert.Equal(t, false, sets.Indexed.GetBool(xKey, xIndex, true))
	sets.Indexed.SetDefault(xKey, "true")
	assert.Equal(t, true, sets.Indexed.GetBool(xKey, xIndex, false))
	sets.Indexed.Set(xKey, xIndex, "false")
	assert.Equal(t, false, sets.Indexed.GetBool(xKey, xIndex, true))

	sets.DumpTo(os.Stdout)

	//Test.Ln()
	//
	//indexKey := "[192.168.11.12:55555]"
	//isVerbose := sets.GetIndexedBool("radar", indexKey, "udp.verbose.trigger", false)
	//Test.Ln("isVerbose", isVerbose)
	//
	//sets.SetRaw("radar.[192.168.11.12:55555].udp.verbose.trigger", "true")
	//isVerbose = sets.GetIndexedBool("radar", indexKey, "udp.verbose.trigger", false)
	//sets.DumpTo(os.Stdout)
	//Test.Ln("isVerbose", isVerbose)
}

func TestSettings_IndexedRepeat(t *testing.T) {
	sets := new(Settings)
	sets.Init()

	sets.Indexed.Set(xKey, "[.11]", "Whatever")
	sets.Indexed.Set(xKey, "[.12]", "Whomever")
	sets.Indexed.Set(xKey, "[.13]", "Whomever")

	keys := sets.Indexed.GetAll(xKey)
	assert.Equal(t, 3, len(keys))

}
