package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"rvpro3/radarvision.com/cmd/stresstool/hive"
	"rvpro3/radarvision.com/utils"
)

const MtBroadcastStatus = 1
const MtBroadcastObjectList = 2

func main() {
	utils.Print.Ln("Radar Vision")
	utils.Print.Ln("RVPro Socket Dump Tool - Copyright Radar Vision 2026")

	wd, err := os.Getwd()
	utils.Debug.Panic(err)

	utils.Print.Ln("Program Directory: ", wd)

	gs := &utils.GlobalSettings
	gs.ReadArgs()

	host := gs.GetBasicDef("socket.host", "")
	switch host {
	case "":
		showHelp()

	default:
		dumpSocket()
	}

	utils.Print.Ln("Done")
}

func showHelp() {
	utils.Print.Ln("Usage:", os.Args[0], "[options]")
	utils.Print.Ln("Options:")
	utils.Print.Option("-o=setting.socket.host")
	utils.Print.Descrp("Host or IP address of RVPro")
	utils.Print.Sample("Sample: -o=setting.socket.host=10.8.1.10")
	utils.Print.Sample("Default: [empty]")

	utils.Print.Option("-o=setting.dump.dir")
	utils.Print.Descrp("Directory where socket output file will be written")
	utils.Print.Sample("Sample: -o=setting.dump.dir=./socket-dump")
	utils.Print.Sample("Default is .")

	utils.Print.Option("-o=setting.dump.file")
	utils.Print.Descrp("File name of socket dump")
	utils.Print.Sample("-o=setting.dump.file=mydump.json")
	utils.Print.Sample("Default: The current date with .json extension")
}

func dumpSocket() {
	gs := &utils.GlobalSettings
	host := gs.GetBasicDef("socket.host", "")
	sock := hive.RVProSocket{}
	sock.Init(fmt.Sprintf("wss://%s/socket", host))

	utils.Print.InfoLn("Connecting to ", sock.Url)

	utils.Debug.Panic(sock.Connect())
	defer sock.Disconnect()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go sock.RunReadSubscription(7, &wg, dumpData)
	wg.Wait()

	if sock.Err != nil {
		utils.Print.ErrorLn("Socket Error:", sock.Err)
	}

	if file != nil {
		_ = writer.Flush()
		_ = file.Close()
	}
}

var file *os.File
var writer *bufio.Writer

func dumpData(bytes []byte) error {
	var err error
	gs := &utils.GlobalSettings

	if file == nil {
		targetDir := gs.GetBasicDef("dump.dir", ".")
		defaultFile := fmt.Sprintf("Dump-%s.json", time.Now().Format(utils.FileDateTimeMS))
		targetFile := gs.GetBasicDef("dump.file", defaultFile)

		fullPath := filepath.Join(targetDir, targetFile)
		utils.Print.InfoLn("New Dump: ", fullPath)

		if file, err = os.Create(fullPath); err != nil {
			return err
		}
		writer = bufio.NewWriter(file)
	}

	messageType := binary.LittleEndian.Uint16(bytes[:2])
	messageSize := binary.LittleEndian.Uint16(bytes[2:4])

	switch messageType {
	case MtBroadcastStatus:
		err = extractStatus(messageSize, bytes[:])

	case MtBroadcastObjectList:
		err = extractObjectList(messageSize, bytes[:])

	default:
		return nil
		// Ignore
		//return errors.Errorf("Unsupported message type %d", messageType)
	}
	err = writer.WriteByte('\n')
	return err
}

func extractStatus(messageSize uint16, bytes []byte) (err error) {
	_, err = writer.Write(bytes[4 : len(bytes)-2])
	return err
}

func extractObjectList(messageSize uint16, bytes []byte) (err error) {
	obj := ObjList{}

	err = obj.Deserialize(bytes)
	if err != nil {
		return err
	}

	var jsonData []byte
	jsonData, err = json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = writer.Write(jsonData)
	return err
}

type ObjList struct {
	ReceivedOn time.Time
	RadarIP    utils.IP4
	ObjCount   uint16
	Details    []objDetail
}

func (o *ObjList) Deserialize(buf []byte) (err error) {
	o.ReceivedOn = time.Now()
	reader := bytes.NewReader(buf)
	_, _ = reader.Seek(4, io.SeekStart)

	_, _ = reader.Read(o.RadarIP.Bytes[:])
	err = binary.Read(reader, binary.LittleEndian, &o.ObjCount)

	o.Details = make([]objDetail, o.ObjCount)

	for i := 0; i < int(o.ObjCount); i++ {
		err = o.Details[i].Deserialize(reader)
	}
	return err
}

type objDetail struct {
	ObjectId    uint16
	ObjectClass uint8
	ClosestLane uint16
	OnZone      uint16
	XFront      float32
	YFront      float32
	Length      float32
	Heading     float32
	Speed       float32
}

func (d *objDetail) Deserialize(reader *bytes.Reader) (err error) {
	err = binary.Read(reader, binary.LittleEndian, &d.ObjectId)
	d.ObjectClass, err = reader.ReadByte()
	err = binary.Read(reader, binary.LittleEndian, &d.ClosestLane)
	err = binary.Read(reader, binary.LittleEndian, &d.OnZone)

	d.XFront, err = d.unpack1(reader)
	d.YFront, err = d.unpack1(reader)
	d.Length, err = d.unpack1(reader)
	d.Heading, err = d.unpack1(reader)
	d.Speed, err = d.unpack1(reader)
	return err
}

func (d *objDetail) unpack1(reader *bytes.Reader) (float32, error) {
	var i16 int16
	err := binary.Read(reader, binary.LittleEndian, &i16)

	value := float64(i16) / math.Pow10(1)
	return float32(value), err

}

//func abc() {
//	b.writeU16(&buffer, uint16(b.MessageType()))
//	b.writeU16(&buffer, sizeFiller)
//
//	objCount := objList.ObjList.Header.NofObjects
//
//	b.writeU32BE(&buffer, objList.RadarIP)
//	b.writeU16(&buffer, objCount)
//
//	for n := 0; n < int(objCount); n++ {
//		obj := &objList.ObjList.Details[n].Payload
//		b.writeU16(&buffer, obj.ObjectId)
//		buffer.WriteByte(uint8(obj.ObjectClass))
//		b.writeU16(&buffer, obj.ClosestLane)
//		b.writeU16(&buffer, uint16(obj.OnZone))
//
//		b.writeI16(&buffer, rvmath.Pack1(obj.XFront))
//		b.writeI16(&buffer, rvmath.Pack1(obj.YFront))
//		b.writeI16(&buffer, rvmath.Pack1(obj.Length))
//		b.writeI16(&buffer, rvmath.Pack1(obj.Heading))
//		b.writeI16(&buffer, rvmath.Pack1(obj.AbsoluteSpeed))
//	}
//
//}
