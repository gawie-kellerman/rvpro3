package port

import (
	"encoding/binary"
	"math"
	"net"
	"strconv"

	"rvpro3/radarvision.com/utils"
)

type InstructionDataType uint8

const (
	None InstructionDataType = iota
	IdtI8
	IdtU8
	IdtI16
	IdtU16
	IdtI32
	IdtU32
	IdtF32
	IdtU64
	IdtF64
)

func (dt InstructionDataType) ToString() string {
	switch dt {
	case IdtI8:
		return "i8"
	case IdtU8:
		return "u8"
	case IdtI16:
		return "i16"
	case IdtU16:
		return "u16"
	case IdtI32:
		return "i32"
	case IdtU32:
		return "u32"
	case IdtF32:
		return "f32"
	case IdtU64:
		return "u64"
	case IdtF64:
		return "f64"
	default:
		return "unknown"
	}
}

type InstructionRequestType uint8

const (
	ReqTypeInvalid InstructionRequestType = iota
	ReqTypeSetParameter
	ReqTypeGetParameter
	ReqTypeReadStatus
	ReqTypeCommand
	ReqTypeExportParamOrStatus

	//Command RequestType = iota
	//StatusRequest
	//ParameterWrite
	//ParameterRead
	//ParameterWriteRead
)

func (irt InstructionRequestType) ToString() string {
	switch irt {
	case ReqTypeInvalid:
		return "invalid"
	case ReqTypeSetParameter:
		return "set parameter"
	case ReqTypeGetParameter:
		return "get parameter"
	case ReqTypeReadStatus:
		return "read status"
	case ReqTypeCommand:
		return "command"
	case ReqTypeExportParamOrStatus:
		return "export param"
	default:
		return "unknown"
	}
}

type InstructionResponseType uint8

const (
	ResTypeNoInstruction InstructionResponseType = iota
	ResTypeSuccess
	ResTypeGeneralError
	ResTypeInvalidRequest
	ResTypeInvalidSection
	InvalidId
	InvalidProtection
	OutOfMinimalBounds
	OutOfMaximalBounds
	ValueIsNotANumber
	InvalidInstruction
	InvalidDimension
	InvalidElement
	InvalidSignature
	InvalidAccessLevel
)

func (rt InstructionResponseType) ToString() string {
	switch rt {
	case ResTypeNoInstruction:
		return "No instruction"
	case ResTypeSuccess:
		return "Success"
	case ResTypeGeneralError:
		return "General Error"
	case ResTypeInvalidRequest:
		return "Invalid Request"
	case ResTypeInvalidSection:
		return "Invalid InstructionSection"
	case InvalidId:
		return "Invalid Id"
	case InvalidProtection:
		return "Invalid Protection"
	case OutOfMinimalBounds:
		return "Out Of Minimal Bounds"
	case OutOfMaximalBounds:
		return "Out Of Maximal Bounds"
	case ValueIsNotANumber:
		return "Value Is Not A Number"
	case InvalidInstruction:
		return "Invalid instruction"
	case InvalidDimension:
		return "Invalid Dimension"
	case InvalidElement:
		return "Invalid Element"
	case InvalidSignature:
		return "Invalid Signature"
	case InvalidAccessLevel:
		return "Invalid AccessLevel"
	default:
		return "Unknown"
	}
}

type InstructionHeader struct {
	NoInstructions uint8
	SequenceNo     uint32
}

func (ih *InstructionHeader) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	ih.NoInstructions = reader.ReadU8()
	reader.SkipRead(3)
	ih.SequenceNo = reader.ReadU32(order)
}

func (ih *InstructionHeader) Write(writer *utils.FixedBuffer, order binary.ByteOrder) {
	writer.WriteU8(ih.NoInstructions)
	writer.SkipWrite(3)
	writer.WriteU32(ih.SequenceNo, order)
}

func (ih *InstructionHeader) PrintDetail() {
	utils.Print.Detail("instruction Header", "\n")
	utils.Print.Indent(2)
	_, _ = utils.Print.Detail("Instructions", "%d\n", ih.NoInstructions)
	_, _ = utils.Print.Detail("Sequence Count", "%d\n", ih.SequenceNo)
	utils.Print.Indent(-2)
}

func (ih *InstructionHeader) GetByteSize() int {
	return 8
}

func (ih *InstructionHeader) GetDetailByteSize() int {
	res := 1 + 1 + 2 + 2 + 1 + 1 + 2 + 2 + 4 + 8
	res *= int(ih.NoInstructions)
	return res
}

type InstructionDetail struct {
	RequestType  InstructionRequestType
	ResponseType InstructionResponseType
	SectionId    uint16
	ParameterId  uint16
	DataType     InstructionDataType
	DimCount     uint8
	Element1     uint16
	Element2     uint16
	Signature    uint32
	Value        [8]byte
}

func (id *InstructionDetail) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	id.RequestType = InstructionRequestType(reader.ReadU8())
	id.ResponseType = InstructionResponseType(reader.ReadU8())
	id.SectionId = reader.ReadU16(order)
	id.ParameterId = reader.ReadU16(order)
	id.DataType = InstructionDataType(reader.ReadU8())
	id.DimCount = reader.ReadU8()
	id.Element1 = reader.ReadU16(order)
	id.Element2 = reader.ReadU16(order)
	id.Signature = reader.ReadU32(order)
	reader.ReadBuffer(id.Value[:])
}

func (id *InstructionDetail) Sign1Dim(sectionName string, parameterName string, dimName string, dimElements int) {
	id.Signature = Calc1Dim(
		int(id.SectionId),
		int(id.ParameterId),
		sectionName,
		parameterName,
		int(id.DimCount),
		dimName,
		dimElements,
		id.DataType.ToString(),
	)
}

func (id *InstructionDetail) Write(writer *utils.FixedBuffer, order binary.ByteOrder) {
	writer.WriteU8(uint8(id.RequestType))
	writer.WriteU8(uint8(id.ResponseType))
	writer.WriteU16(id.SectionId, order)
	writer.WriteU16(id.ParameterId, order)
	writer.WriteU8(uint8(id.DataType))
	writer.WriteU8(id.DimCount)
	writer.WriteU16(id.Element1, order)
	writer.WriteU16(id.Element2, order)
	writer.WriteU32(id.Signature, order)
	writer.WriteBytes(id.Value[:])
}

func (id *InstructionDetail) GetF32(order binary.ByteOrder) float32 {
	bits := order.Uint32(id.Value[:])
	return math.Float32frombits(bits)
}

func (id *InstructionDetail) GetF64Value(order binary.ByteOrder) float64 {
	bits := order.Uint64(id.Value[:])
	return math.Float64frombits(bits)
}

func (id *InstructionDetail) Sign(sectionName string, parameterName string) {
	step1 := sectionName
	step1 += strconv.FormatUint(uint64(id.SectionId), 10)
	step1 += strconv.FormatUint(uint64(id.DimCount), 10)
	step1Crc := CalcCRC32(step1)

	step2 := parameterName
	step2 += strconv.FormatUint(uint64(id.ParameterId), 10)
	step2 += id.DataType.ToString()
	step2 += strconv.FormatUint(uint64(step1Crc), 10)
	step2 += strconv.FormatUint(uint64(id.DimCount), 10)

	id.Signature = CalcCRC32(step2)
}

func (id *InstructionDetail) PrintDetail(no int, order binary.ByteOrder) {
	utils.Print.Detail("instruction Detail", "# %d\n", no)
	utils.Print.Indent(2)
	utils.Print.Detail("Request Type", "%d, %s\n", id.RequestType, id.RequestType.ToString())
	utils.Print.Detail("Response Type", "%d, %s\n", id.ResponseType, id.ResponseType.ToString())
	utils.Print.Detail("InstructionSection", "%d\n", id.SectionId)
	utils.Print.Detail("Parameter", "%d\n", id.ParameterId)
	utils.Print.Detail("DataType", "%d, %s\n", id.DataType, id.DataType.ToString())
	utils.Print.Detail("Dim Count", "%d\n", id.DimCount)
	utils.Print.Detail("Element", "%d, %d\n", id.Element1, id.Element2)
	utils.Print.Detail("Signature", "%d\n", id.Signature)
	utils.Print.Detail("Value", "%s\n", id.ToString(order))
	utils.Print.Indent(-2)
}

func (id *InstructionDetail) ToString(order binary.ByteOrder) string {
	switch id.DataType {
	case IdtF32:
		return strconv.FormatFloat(float64(id.GetF32(order)), 'f', -1, 32)
	case IdtF64:
		return strconv.FormatFloat(id.GetF64Value(order), 'f', -1, 64)

	case IdtI32:
		return strconv.Itoa(int(id.GetI32(order)))

	case IdtU32:
		return strconv.Itoa(int(id.GetU32(order)))

	case IdtI8:
		return strconv.Itoa(int(id.GetI8()))

	case IdtU8:
		return strconv.Itoa(int(id.GetU8()))

	case IdtI16:
		return strconv.Itoa(int(id.GetI16(order)))

	case IdtU16:
		return strconv.Itoa(int(id.GetU16(order)))
	}
	return "unmapped"
}

func (id *InstructionDetail) GetI32(order binary.ByteOrder) int32 {
	bits := order.Uint32(id.Value[:])
	return int32(bits)
}

func (id *InstructionDetail) GetU32(order binary.ByteOrder) uint32 {
	bits := order.Uint32(id.Value[:])
	return bits
}

func (id *InstructionDetail) GetI8() int8 {
	return int8(id.Value[0])
}

func (id *InstructionDetail) GetU8() uint8 {
	return id.Value[0]
}

func (id *InstructionDetail) GetI16(order binary.ByteOrder) int16 {
	bits := order.Uint16(id.Value[:])
	return int16(bits)
}

func (id *InstructionDetail) GetU16(order binary.ByteOrder) uint16 {
	bits := order.Uint16(id.Value[:])
	return bits
}

type Instruction struct {
	Th       TransportHeader
	Ph       PortHeader
	Header   InstructionHeader
	Detail   []InstructionDetail
	Crc      uint16
	CrcCheck uint16
}

func NewInstruction() *Instruction {
	res := &Instruction{}
	res.Th.Init()
	res.Ph.Init(PiInstruction)
	res.Detail = make([]InstructionDetail, 0, 1)
	return res
}

func (i *Instruction) Read(reader *utils.FixedBuffer) error {
	i.Th.Read(reader)
	reader.StartReadMarker()
	i.Ph.Read(reader)

	if reader.Err != nil {
		return reader.Err
	}

	order := i.Ph.GetOrder()
	i.Header.Read(reader, order)
	i.Detail = make([]InstructionDetail, i.Header.NoInstructions)

	for n := 0; n < len(i.Detail); n++ {
		detail := &i.Detail[n]
		detail.Read(reader, order)
	}

	if !i.Th.Flags.IsSkipPayloadCrc() {
		i.CrcCheck = reader.CalcReadCRC()
		i.Crc = reader.ReadU16(binary.BigEndian)
	}

	if reader.Err != nil {
		return reader.Err
	}

	return i.Validate()
}

func (i *Instruction) Validate() (err error) {
	if err = i.Th.Validate(); err != nil {
		return err
	}

	if i.Crc != i.CrcCheck {
		return ErrPayloadCRC
	}

	return nil
}

func (i *Instruction) Write(writer *utils.FixedBuffer) error {
	writer.StartWriteMarker()
	i.Th.PayloadLength = i.GetPayloadSize()
	i.Th.Write(writer)
	i.Th.CRC16 = writer.WriteCRC16(binary.BigEndian)
	i.Th.CheckCRC16 = i.Th.CRC16

	writer.StartWriteMarker()
	order := i.Ph.GetOrder()

	i.Ph.PortSize = uint32(i.Th.PayloadLength)
	i.Ph.Write(writer)

	if writer.Err != nil {
		return writer.Err
	}

	i.Header.Write(writer, order)
	for n := 0; n < len(i.Detail); n++ {
		detail := &i.Detail[n]
		detail.Write(writer, order)
	}

	if !i.Th.Flags.IsSkipPayloadCrc() {
		i.CrcCheck = writer.CalcWriteCRC()
		i.Crc = i.CrcCheck
		writer.WriteU16(i.Crc, binary.BigEndian)
	}

	return writer.Err
}

func (i *Instruction) WriteToUDP(ip4 utils.IP4) (err error) {
	var udpConn *net.UDPConn
	var buffer [4 * utils.Kilobyte]byte
	writer := utils.NewFixedBuffer(buffer[:], 0, 0)

	if err = i.Write(&writer); err != nil {
		return err
	}

	remoteAddr := ip4.ToUDPAddr()
	if udpConn, err = net.DialUDP("udp4", nil, &remoteAddr); err != nil {
		return err
	}
	defer udpConn.Close()

	_, err = udpConn.Write(writer.AsWriteSlice())
	return err
}

func (i *Instruction) PrintDetail() {
	i.Th.PrintDetail()
	i.Ph.PrintDetail()
	i.Header.PrintDetail()
	for index := range i.Detail {
		detail := &i.Detail[index]
		detail.PrintDetail(index, i.Ph.GetOrder())
	}
}

func (i *Instruction) AddDetail(detail InstructionDetail) *InstructionDetail {
	if i.Header.NoInstructions == 0 {
		i.Detail = make([]InstructionDetail, 0, 2)
	}

	i.Header.NoInstructions++
	i.Detail = append(i.Detail, detail)

	res := &i.Detail[len(i.Detail)-1]
	*res = detail
	return res
}

func (i *Instruction) GetPayloadSize() uint16 {
	res := 0
	res += i.Ph.GetByteSize()
	res += i.Header.GetByteSize()
	res += i.Header.GetDetailByteSize()
	return uint16(res)
}

func (i *Instruction) SaveAsBytes() []byte {
	var res []byte
	res = make([]byte, i.GetTotalSize())

	writer := utils.NewFixedBuffer(res, 0, 0)
	if err := i.Write(&writer); err != nil {
		panic(err)
	}

	return writer.AsWriteSlice()
}

func (i *Instruction) GetTotalSize() int {
	return int(i.Th.GetSize()) + int(i.GetPayloadSize()) + 2
}

func (i *Instruction) GetDetail(index int) *InstructionDetail {
	return &i.Detail[index]
}
