package port

import (
	"hash/crc32"
	"strconv"
)

func Calc1Dim(
	sectionId int,
	parameterId int,
	sectionName string,
	parameterName string,
	dimCount int,
	dim1Name string,
	dim1Elems int,
	typeStr string,
) uint32 {
	step1 := sectionName
	step1 += strconv.FormatUint(uint64(sectionId), 10)
	step1 += strconv.FormatUint(uint64(dimCount), 10)
	step1Crc := CalcCRC32(step1)

	stepD := dim1Name + strconv.FormatUint(uint64(dim1Elems), 10)
	stepDCrc := CalcCRC32(stepD)

	step2 := parameterName
	step2 += strconv.FormatUint(uint64(parameterId), 10)
	step2 += typeStr
	step2 += strconv.FormatUint(uint64(step1Crc), 10)
	step2 += strconv.FormatUint(uint64(stepDCrc), 10)

	return CalcCRC32(step2)
}

func Calc2Dim(
	sectionId int,
	parameterId int,
	sectionName string,
	parameterName string,
	dimCount int,
	dim1Name string,
	dim1Elems int,
	dim2Name string,
	dim2Elems int,
	typeStr string,
) uint32 {
	step1 := sectionName
	step1 += strconv.FormatUint(uint64(sectionId), 10)
	step1 += strconv.FormatUint(uint64(dimCount), 10)
	step1Crc := CalcCRC32(step1)

	stepD := dim1Name + strconv.FormatUint(uint64(dim1Elems), 10)
	stepD += dim2Name + strconv.FormatUint(uint64(dim2Elems), 10)
	stepDCrc := CalcCRC32(stepD)

	step2 := parameterName
	step2 += strconv.FormatUint(uint64(parameterId), 10)
	step2 += typeStr
	step2 += strconv.FormatUint(uint64(step1Crc), 10)
	step2 += strconv.FormatUint(uint64(stepDCrc), 10)

	return CalcCRC32(step2)
}

func CalcCRC32(source string) uint32 {
	size := len(source) * 2
	barr := make([]byte, size)

	for n := 0; n < len(source); n++ {
		barr[n*2+1] = source[n]
	}

	return crc32.ChecksumIEEE(barr)
}
