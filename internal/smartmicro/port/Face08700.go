package port

type face08700 struct {
	Zones        face08700AppTmZones
	ZoneSegments face08700AppTmZoneSegments
	Parameters   face08700AppParameters
}

var Face08700 face08700

type face08700AppTmZones struct {
}

type face08700AppTmZoneSegments struct{}
type face08700AppParameters struct{}

func (face08700AppTmZoneSegments) instruction(paramId int, paramName string, dt InstructionDataType, element1 int) InstructionDetail {
	res := InstructionDetail{
		SectionId:    3019,
		ParameterId:  uint16(paramId),
		DimCount:     1,
		RequestType:  ReqTypeGetParameter,
		ResponseType: ResTypeNoInstruction,
		DataType:     dt,
		Element1:     uint16(element1),
		Element2:     0,
	}

	res.Signature = Calc1Dim(
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

func (face08700AppTmZones) instruction(paramId int, paramName string, dt InstructionDataType, element1 int) InstructionDetail {
	res := InstructionDetail{
		SectionId:    3018,
		ParameterId:  uint16(paramId),
		DimCount:     1,
		RequestType:  ReqTypeGetParameter,
		ResponseType: ResTypeNoInstruction,
		DataType:     dt,
		Element1:     uint16(element1),
		Element2:     0,
	}

	res.Signature = Calc1Dim(
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

func (s face08700AppTmZones) GetRelayAssignment(zone int) InstructionDetail {
	return s.instruction(2, "relay_assignment", IdtU8, zone)
}

func (s face08700AppTmZones) GetNofSegmentsByZone(zone int) InstructionDetail {
	return s.instruction(0, "used_segments", IdtU8, zone)
}

func (s face08700AppTmZones) GetWidthByZone(zone int) InstructionDetail {
	return s.instruction(4, "zone_width", IdtF32, zone)
}

func (s face08700AppTmZones) IsNofSegmentsByZone(detail *InstructionDetail) bool {
	return detail.SectionId == 3018 && detail.ParameterId == uint16(0)
}

func (s face08700AppTmZones) IsWidthByZone(detail *InstructionDetail) bool {
	return detail.SectionId == 3018 && detail.ParameterId == uint16(4)
}

func (s face08700AppTmZones) IsRelayAssignment(detail *InstructionDetail) bool {
	return detail.SectionId == 3018 && detail.ParameterId == uint16(2)
}

func (s face08700AppTmZoneSegments) GetXSegment(segment int) InstructionDetail {
	return s.instruction(0, "pos_x", IdtF32, segment)
}

func (s face08700AppTmZoneSegments) GetYSegment(segment int) InstructionDetail {
	return s.instruction(1, "pos_y", IdtF32, segment)
}

func (s face08700AppTmZoneSegments) IsGetXSegment(detail *InstructionDetail) bool {
	return detail.SectionId == 3019 && detail.ParameterId == uint16(0)
}

func (s face08700AppTmZoneSegments) IsGetYSegment(detail *InstructionDetail) bool {
	return detail.SectionId == 3019 && detail.ParameterId == uint16(1)
}

func (s face08700AppParameters) instruction(paramId int, paramName string, dt InstructionDataType) InstructionDetail {
	res := InstructionDetail{
		SectionId:    3017,
		ParameterId:  uint16(paramId),
		DimCount:     0,
		RequestType:  ReqTypeGetParameter,
		ResponseType: ResTypeNoInstruction,
		DataType:     dt,
		Element1:     0,
		Element2:     0,
	}
	res.Sign("app_tm_parameters", paramName)
	return res
}

func (s face08700AppParameters) GetNofZones() InstructionDetail {
	return s.instruction(0, "nof_zones", IdtU16)
}

func (s face08700AppParameters) IsGetNofZones(detail *InstructionDetail) bool {
	return detail.SectionId == 3017 && detail.ParameterId == uint16(0)
}
