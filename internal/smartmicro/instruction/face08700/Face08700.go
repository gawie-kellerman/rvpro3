package face08700

import (
	"encoding/binary"

	"rvpro3/radarvision.com/internal/smartmicro/port"
)

type detail struct {
	Zones        appTmZones
	ZoneSegments appTmZoneSegments
	Parameters   appParameters
}

var Detail detail

type appTmZones struct{}

type appTmZoneSegments struct{}
type appParameters struct{}

const ParameterSimulationMode = 3

func (appTmZoneSegments) instruction(paramId int, paramName string, dt port.InstructionDataType, element1 int) port.InstructionDetail {
	res := port.InstructionDetail{
		SectionId:    3019,
		ParameterId:  uint16(paramId),
		DimCount:     1,
		RequestType:  port.ReqTypeGetParameter,
		ResponseType: port.ResTypeNoInstruction,
		DataType:     dt,
		Element1:     uint16(element1),
		Element2:     0,
	}

	res.Signature = port.Calc1Dim(
		int(res.SectionId),
		int(res.ParameterId),
		"app_tm_zone_segments",
		paramName,
		1,
		"MAX_NOF_ZONE_SEGMENTS",
		128,
		dt.ToString(),
	)
	return res
}

func (appTmZones) instruction(paramId int, paramName string, dt port.InstructionDataType, element1 int) port.InstructionDetail {
	res := port.InstructionDetail{
		SectionId:    3018,
		ParameterId:  uint16(paramId),
		DimCount:     1,
		RequestType:  port.ReqTypeGetParameter,
		ResponseType: port.ResTypeNoInstruction,
		DataType:     dt,
		Element1:     uint16(element1),
		Element2:     0,
	}

	res.Signature = port.Calc1Dim(
		int(res.SectionId),
		int(res.ParameterId),
		"app_tm_zones",
		paramName,
		1,
		"MAX_NOF_ZONES",
		32,
		dt.ToString(),
	)
	return res
}

func (s appTmZones) GetRelayAssignment(zone int) port.InstructionDetail {
	return s.instruction(2, "relay_assignment", port.IdtU8, zone)
}

func (s appTmZones) GetNofSegmentsByZone(zone int) port.InstructionDetail {
	return s.instruction(0, "used_segments", port.IdtU8, zone)
}

func (s appTmZones) GetWidthByZone(zone int) port.InstructionDetail {
	return s.instruction(4, "zone_width", port.IdtF32, zone)
}

func (s appTmZones) IsNofSegmentsByZone(detail *port.InstructionDetail) bool {
	return detail.SectionId == 3018 && detail.ParameterId == uint16(0)
}

func (s appTmZones) IsWidthByZone(detail *port.InstructionDetail) bool {
	return detail.SectionId == 3018 && detail.ParameterId == uint16(4)
}

func (s appTmZones) IsRelayAssignment(detail *port.InstructionDetail) bool {
	return detail.SectionId == 3018 && detail.ParameterId == uint16(2)
}

func (s appTmZoneSegments) GetXSegment(segment int) port.InstructionDetail {
	return s.instruction(0, "pos_x", port.IdtF32, segment)
}

func (s appTmZoneSegments) GetYSegment(segment int) port.InstructionDetail {
	return s.instruction(1, "pos_y", port.IdtF32, segment)
}

func (s appTmZoneSegments) IsGetXSegment(detail *port.InstructionDetail) bool {
	return detail.SectionId == 3019 && detail.ParameterId == uint16(0)
}

func (s appTmZoneSegments) IsGetYSegment(detail *port.InstructionDetail) bool {
	return detail.SectionId == 3019 && detail.ParameterId == uint16(1)
}

func (s appParameters) instruction(paramId int, paramName string, dt port.InstructionDataType) port.InstructionDetail {
	res := port.InstructionDetail{
		SectionId:    AppTMParametersSection,
		ParameterId:  uint16(paramId),
		DimCount:     0,
		RequestType:  port.ReqTypeGetParameter,
		ResponseType: port.ResTypeNoInstruction,
		DataType:     dt,
		Element1:     0,
		Element2:     0,
	}
	res.Sign("app_tm_parameters", paramName)
	return res
}

func (s appParameters) GetNofZones() port.InstructionDetail {
	return s.instruction(0, "nof_zones", port.IdtU16)
}

func (appParameters) IsGetNofZones(detail *port.InstructionDetail) bool {
	return detail.SectionId == 3017 && detail.ParameterId == uint16(0)
}

func (s appParameters) GetSimulationMode() port.InstructionDetail {
	return s.instruction(S3017SimulationMode, S3017SimulationModeName, port.IdtU16)
}

func (s appParameters) SetSimulationMode(order binary.ByteOrder, value uint16) port.InstructionDetail {
	res := s.instruction(S3017SimulationMode, S3017SimulationModeName, port.IdtU16)
	res.RequestType = port.ReqTypeSetParameter
	res.SetU16(order, value)
	return res
}
