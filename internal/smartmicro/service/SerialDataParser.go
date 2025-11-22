package service

import (
	"encoding/binary"
	"log"

	"rvpro3/radarvision.com/internal/smartmicro/port"
	"rvpro3/radarvision.com/utils"
)

type SerialDataParser struct {
}

const headerOffset = 26
const payloadStart = 3
const payloadEnd = 5

// Parse a buffer.  The buffer is the slice to the full
// current buffer
func (s *SerialDataParser) Parse(buffer []byte) []byte {
	slice := buffer

	for {
		slice = s.cutFront(slice)
		if len(slice) > 10 {
			headerLen := int(slice[headerOffset])
			payloadLen := int(binary.BigEndian.Uint16(slice[payloadStart:payloadEnd]))
			totalLen := headerLen + payloadLen

			if len(slice) >= totalLen {
				slice = s.parseMessage(slice)
			} else {
				break
			}
		} else {
			break
		}
	}

	return slice
}

func (s *SerialDataParser) cutFront(slice []byte) []byte {
	for len(slice) > 0 {
		if slice[0] == port.StartPattern {
			break
		}
		slice = slice[1:]
	}
	return slice
}

func (s *SerialDataParser) parseMessage(buffer []byte) []byte {
	var err error
	reader := utils.NewFixedBuffer(buffer, 0, len(buffer))
	var th port.TransportHeader
	//var ph port.PortHeader

	th.Read(&reader)

	if reader.Err != nil {
		err = reader.Err
		goto errorLabel
	}

	if err = th.Validate(); err != nil {
		goto errorLabel
	}

	//ph.ReadPortData(&reader)
	reader.SeekRead(int(th.PayloadLength))

	return reader.AvailForRead()

errorLabel:
	log.Fatal(err)
	return reader.AvailForRead()
}
