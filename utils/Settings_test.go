package utils

import (
	"os"
	"testing"
)

func TestSettings_DumpTo(t *testing.T) {
	sets := &Settings{}
	sets.Init()

	sets.SetDefault("radar.udp.verbose.trigger", "false")
	sets.DumpTo(os.Stdout)

	Test.Ln()

	indexKey := "[192.168.11.12:55555]"
	isVerbose := sets.GetIndexedAsBool("radar", indexKey, "udp.verbose.trigger")
	Test.Ln("isVerbose", isVerbose)

	sets.SetRaw("radar.[192.168.11.12:55555].udp.verbose.trigger", "true")
	isVerbose = sets.GetIndexedAsBool("radar", indexKey, "udp.verbose.trigger")
	sets.DumpTo(os.Stdout)
	Test.Ln("isVerbose", isVerbose)
}
