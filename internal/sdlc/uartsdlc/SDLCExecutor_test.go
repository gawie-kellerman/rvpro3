package uartsdlc

import (
	"fmt"
	"testing"

	"rvpro3/radarvision.com/utils"
)

func TestSdlcExecutorSettings_Setup(t *testing.T) {
	exec := SDLCExecutorSettings.SetupNew(&utils.GlobalConfig)
	fmt.Println(exec)
}
