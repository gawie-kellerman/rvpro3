package main

import (
	"os"
	"testing"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

func TestLoadInstruction(t *testing.T) {
	//bytes, err := os.ReadFile("/home/gkellerman/2-to-12.instruction.bin")
	bytes, err := os.ReadFile("/home/gkellerman/2-to-12.instruction-5.bin")
	//bytes, err := os.ReadFile("/home/gkellerman/Workspace/RadarVision/Source/rvpro3/cmd/util/radar-startsim/instruction.bin")
	utils.Debug.Panic(err)

	ins := port.Instruction{}
	err = ins.ReadBytes(bytes)
	utils.Debug.Panic(err)
}
