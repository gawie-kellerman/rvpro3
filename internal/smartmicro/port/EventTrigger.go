package port

import (
	"encoding/binary"
	"os"

	"rvpro3/radarvision.com/utils"
)

type EventTrigger struct {
	Th       TransportHeader
	Ph       PortHeader
	Header   EventTriggerHeader
	Crc      uint16
	CrcCheck uint16
}

type EventTriggerHeader struct {
	reserved            uint8
	NofTriggeredRelays  uint8
	NofTriggeredObjects uint8
	FeatureFlags        uint8
	Relays1             uint32
	Relays2             uint32
}

func (s *EventTriggerHeader) Read(reader *utils.FixedBuffer, order binary.ByteOrder) {
	s.reserved = reader.ReadU8()
	s.NofTriggeredRelays = reader.ReadU8()
	s.NofTriggeredObjects = reader.ReadU8()
	s.FeatureFlags = reader.ReadU8()
	s.Relays1 = reader.ReadU32(order)
	s.Relays2 = reader.ReadU32(order)
}

func (s *EventTrigger) ReadPortData(reader *utils.FixedBuffer) {
	order := s.Ph.GetOrder()
	reader.StartReadMarker()

	s.Header.Read(reader, order)

	if !s.Th.Flags.IsSkipPayloadCrc() {
		s.CrcCheck = reader.CalcReadCRC()
		s.Crc = reader.ReadU16(binary.BigEndian)
	}
}

func (s *EventTrigger) ReadBytes(bytes []byte) error {
	reader := utils.NewFixedBuffer(bytes, 0, len(bytes))
	s.Th.Read(&reader)
	s.Ph.Read(&reader)
	if reader.Err != nil {
		return reader.Err
	}
	s.ReadPortData(&reader)
	return reader.Err
}

func (s *EventTrigger) ReadFile(filename string) (err error) {
	var data []byte

	if data, err = os.ReadFile(filename); err != nil {
		return err
	}

	if err = s.ReadBytes(data); err != nil {
		return err
	}

	return s.Validate()
}

func (s *EventTrigger) Validate() (err error) {
	if err = s.Th.Validate(); err != nil {
		return err
	}

	if err = s.Ph.Validate(); err != nil {
		return err
	}

	if s.Crc != s.CrcCheck {
		return ErrPayloadCRC
	}

	return nil
}
