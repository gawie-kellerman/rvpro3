package port

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"rvpro3/radarvision.com/internal/smartmicro/port"
)

func TestTransportHeaderReader_CalcCRC16(t *testing.T) {
	buffer, err := os.ReadFile(getT45Path("obj_list/3_0/2_0.seg#2.bin"))
	//buffer, err := os.ReadFile(getT45Path("obj_list/3_0/2_0.seg#1.bin"))
	//buffer, err := os.ReadFile(getT45Path("obj_list/3_0/2_0.noseg#1.bin"))
	panicIf(err)

	th := port.TransportHeaderReader{
		Buffer: buffer,
	}

	panicIf(th.Check())
	th.PrintDetail()

	ph := port.PortHeaderReader{
		Buffer:      buffer,
		StartOffset: int(th.GetHeaderLength()),
	}

	panicIf(ph.Check())
	ph.PrintDetail()

	objList := port.ObjectListReader{}
	objList.Init(buffer)
	objList.PrintDetail()
}

func TestStitching(t *testing.T) {
	file1, err := os.ReadFile(getT45Path("obj_list/3_0/2_0.seg#1.bin"))
	panicIf(err)

	file2, err := os.ReadFile(getT45Path("obj_list/3_0/2_0.seg#2.bin"))
	panicIf(err)

	aggrTh := port.TransportHeaderReader{
		Buffer: file2,
	}

	file3 := aggrTh.CombineTo(file1)

	th := port.TransportHeaderReader{
		Buffer: file3,
	}

	ph := port.PortHeaderReader{
		Buffer:      file3,
		StartOffset: int(th.GetHeaderLength()),
	}

	panicIf(ph.Check())
	ph.PrintDetail()

	objList := port.ObjectListReader{}
	objList.Init(file3)
	objList.PrintDetail()
}

func TestEventTriggerReader(t *testing.T) {
	var th port.TransportHeaderReader
	var ph port.PortHeaderReader
	buffer, err := os.ReadFile(getT45Path("evt_trigger/4_0/2_0#1.bin"))
	panicIf(err)

	evtTrigger := port.EventTriggerReader{}
	evtTrigger.Init(buffer)

	panicIf(evtTrigger.InitTransport(&th))
	panicIf(evtTrigger.InitPort(&ph))

	th.PrintDetail()
	ph.PrintDetail()
	evtTrigger.PrintDetail()
	fmt.Printf("Loaded %d, Parsed %d\n", len(buffer), evtTrigger.TotalSize())
}

func TestPVRReader(t *testing.T) {
	var th port.TransportHeaderReader
	var ph port.PortHeaderReader
	buffer, err := os.ReadFile(getT45Path("pvr/2_0/2.0#1.bin"))
	panicIf(err)

	pvr := port.PVRReader{}
	pvr.Init(buffer)

	panicIf(pvr.InitTransport(&th))
	panicIf(pvr.InitPort(&ph))

	th.PrintDetail()
	ph.PrintDetail()

	pvr.PrintDetail()

	fmt.Printf("Loaded %d, Parsed %d\n", len(buffer), pvr.TotalSize())
}

func TestStatistics(t *testing.T) {
	var th port.TransportHeaderReader
	var ph port.PortHeaderReader
	buffer, err := os.ReadFile(getT45Path("statistics/4_0/2_0#2.bin"))
	panicIf(err)

	stats := port.StatisticsReader{}
	stats.Init(buffer)

	panicIf(stats.InitTransport(&th))
	panicIf(stats.InitPort(&ph))

	th.PrintDetail()
	ph.PrintDetail()
	stats.PrintDetail()
	fmt.Printf("Loaded %d, Parsed %d\n", len(buffer), stats.TotalSize())
}

func TestObjectListReader(t *testing.T) {
	var th port.TransportHeaderReader
	var ph port.PortHeaderReader
	buffer, err := os.ReadFile(getT45Path("obj_list/3_0/2_0.noseg#1.bin"))
	panicIf(err)

	objList := port.ObjectListReader{}
	objList.Init(buffer)

	panicIf(objList.InitTransport(&th))
	panicIf(objList.InitPort(&ph))

	th.PrintDetail()
	ph.PrintDetail()
	objList.PrintDetail()

	fmt.Printf("Loaded %d, Parsed %d\n", len(buffer), objList.TotalSize())
}

func TestObjListDetail(t *testing.T) {
	var bytes []byte
	bytes = []byte("😀hello")
	fmt.Println(len(bytes))
	fmt.Println(hex.EncodeToString(bytes))

}
