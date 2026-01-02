package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/internal/smartmicro/service"
	"rvpro3/radarvision.com/utils"
)

func getZoneWidths() *port.Instruction {
	res := port.NewInstruction()

	res.Th.Flags = res.Th.Flags.Set(port.FlSourceClientId)
	res.Th.SourceClientId = 0x1000001
	res.Ph.Timestamp = 41062
	//res.Header.SequenceNo = 1
	//res.AddDetail(iFace.Detail(face8700.S3017.Id, face8700.S3017NofZones))
	//detail := res.AddDetail(iFace.Detail(face8700.S3018.Id, face8700.S3018ZoneWidth))
	//detail.DimCount = 1
	//detail.Signature = port.Calc1Dim(
	//	int(face8700.S3018.Id),
	//	face8700.S3018ZoneWidth,
	//	face8700.S3018.Name,
	//	"zone_width",
	//	1,
	//	"MAX_NOF_ZONES",
	//	32,
	//	port.IdtF32.String(),
	//)

	//res.AddDetail(port.Face08700.GetZoneWidth(1))
	for n := 0; n < 10; n++ {
		res.AddDetail(port.Face08700.ZoneSegments.GetXSegment(n))
		//res.AddDetail(port.Face08700.Zones.GetWidthByZone(n))
	}
	//res.AddDetail(port.Face08700.GetNofZones())
	//res.AddDetail(port.Face08700.GetNofSegmentsByZone(1))
	//res.AddDetail(port.Face08700.GetZoneWidth(2))
	//res.AddDetail(port.Face08700.Parameters.GetNofZones())
	return res

}

//	func getZoneWidthsInstruction() *port.instruction {
//		iFace := global.GlobalInstructionFaces.Default
//
//		res := port.NewInstruction()
//
//		res.Th.Flags = res.Th.Flags.SetRaw(port.FlSourceClientId)
//		res.Th.SourceClientId = 0x1000001
//		res.Ph.Timestamp = 41062
//		res.AddDetail(iFace.Detail(face8700.S3018.Id, face8700.S3018ZoneWidth))
//
//		return res
//	}
//
// Check to see if you get a response when not sending alive, but sending a transaction
func checkResponse() {
	hostIP := utils.IP4Builder.FromString("192.168.11.1:55555")

	udpAlive := service.NewT44KeepAliveService()
	udpAlive.Init()
	udpAlive.Start(hostIP)

	udpListener := service.UDPData{}
	udpListener.OnData = func(dataService *service.UDPData, addr net.UDPAddr, bytes []byte) {
		//fmt.Println("Received data: ", len(bytes), addr.String())

		reader := utils.NewFixedBuffer(bytes, 0, len(bytes))
		var th port.TransportHeader
		var ph port.PortHeader

		th.Read(&reader)
		ph.Read(&reader)

		if ph.Identifier == port.PiInstruction {
			reader.ResetTo(0, len(bytes))
			var ins port.Instruction
			panicOnError(ins.Read(&reader))
			for i := range ins.Header.NoInstructions {
				detail := ins.GetDetail(int(i))
				fmt.Printf(
					"Element: %d:%d, status: %s, value: %s of %s\n",
					detail.Element1,
					detail.Element2,
					detail.ResponseType.ToString(),
					detail.ToString(ph.GetOrder()),
					detail.DataType.ToString(),
				)
			}

		}
	}
	udpListener.Start(hostIP)

	time.Sleep(100 * time.Millisecond)
	ins := getZoneWidths()

	ip13 := utils.IP4Builder.FromString("192.168.11.13:55555")
	udpListener.WriteData(ip13, ins.SaveAsBytes())
	time.Sleep(100 * time.Second)

	udpListener.Stop()
	udpAlive.Stop()
}

func loadTest() {
	var err error
	var dir string
	var content []byte

	//makeSimilar()

	dir, err = os.Getwd()
	panicOnError(err)

	fmt.Printf("dir: %s\n", dir)
	fmt.Println("dir:", dir)

	content = loadHexFile("_etc/samples/port/t44-instruction-request.hex")
	reader := utils.NewFixedBuffer(content, 0, len(content))
	ins := port.Instruction{}
	panicOnError(ins.Read(&reader))
	ins.PrintDetail()
}

func main() {
	checkResponse()
	//th, ph := readHeaders(&reader)
	//th.PrintDetail()
	//ph.PrintDetail()
	//ins.Th.PrintDetail()
	//ins.Ph.PrintDetail()
	//ins.Header.PrintDetail()
}

func readHeaders(reader *utils.FixedBuffer) (th port.TransportHeader, ph port.PortHeader) {
	th.Read(reader)
	utils.Print.Detail("Transport Header @", "%d\n", reader.ReadPos)
	ph.Read(reader)
	utils.Print.Detail("Payload Header @", "%d\n", reader.ReadPos)
	panicOnError(reader.Err)
	return th, ph
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func loadHexFile(filename string) []byte {
	_, _ = utils.Print.Detail("Loading hex file", "%s\n", filename)
	content, err := utils.File.LoadFromHex(filename)
	panicOnError(err)

	_, _ = utils.Print.Detail("Binary length", "%d\n", len(content))
	return content
}
