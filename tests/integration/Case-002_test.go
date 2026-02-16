package integration

import (
	"os"
	"testing"

	"rvpro3/radarvision.com/internal/api/radar"
	"rvpro3/radarvision.com/utils"
)

func TestSendTrigger(t *testing.T) {
	//apiUrlStr := getAPIUrl()
	rvmIPStr := getRVMUrl()
	radarIPStr := getRadarIP(1)
	radarIP := utils.IP4Builder.FromString(radarIPStr)

	//var err error

	radarSim := radar.RadarSimulator{
		RadarIP4:  radarIP,
		ServerIP4: utils.IP4Builder.FromString(rvmIPStr),
	}
	defer radarSim.Close()

	dir, err := os.Getwd()
	utils.Debug.Panic(err)

	utils.Test.Ln(dir)
	utils.Debug.Panic(radarSim.SendFile("Trigger-001-4.0.bin"))
}

func TestSendSegregatedObjectList(t *testing.T) {
	//apiUrlStr := getAPIUrl()
	rvmIPStr := getRVMUrl()
	radarIPStr := getRadarIP(1)
	radarIP := utils.IP4Builder.FromString(radarIPStr)

	//var err error

	radarSim := radar.RadarSimulator{
		RadarIP4:  radarIP,
		ServerIP4: utils.IP4Builder.FromString(rvmIPStr),
	}
	defer radarSim.Close()

	dir, err := os.Getwd()
	utils.Debug.Panic(err)

	utils.Test.Ln(dir)
	utils.Debug.Panic(radarSim.SendFile("SegObjList-01-3.0.bin"))
	utils.Debug.Panic(radarSim.SendFile("SegObjList-02-3.0.bin"))
}

func TestSendStats(t *testing.T) {
	//apiUrlStr := getAPIUrl()
	rvmIPStr := getRVMUrl()
	radarIPStr := getRadarIP(1)
	radarIP := utils.IP4Builder.FromString(radarIPStr)

	//var err error

	radarSim := radar.RadarSimulator{
		RadarIP4:  radarIP,
		ServerIP4: utils.IP4Builder.FromString(rvmIPStr),
	}
	defer radarSim.Close()

	dir, err := os.Getwd()
	utils.Debug.Panic(err)

	utils.Test.Ln(dir)
	utils.Debug.Panic(radarSim.SendFile("Statistics-01-4.0.bin"))
}

func TestSendPVR(t *testing.T) {
	//apiUrlStr := getAPIUrl()
	rvmIPStr := getRVMUrl()
	radarIPStr := getRadarIP(1)
	radarIP := utils.IP4Builder.FromString(radarIPStr)

	//var err error

	radarSim := radar.RadarSimulator{
		RadarIP4:  radarIP,
		ServerIP4: utils.IP4Builder.FromString(rvmIPStr),
	}
	defer radarSim.Close()

	dir, err := os.Getwd()
	utils.Debug.Panic(err)

	utils.Test.Ln(dir)
	utils.Debug.Panic(radarSim.SendFile("PVR-01-2.0.bin"))
}
