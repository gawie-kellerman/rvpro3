package uartsdlc

type SDLCIdentifier uint8

const StaticStatusResponseCode SDLCIdentifier = 0x40
const CMUFrameStreamCode SDLCIdentifier = 0x41
const DateTimeStreamCode SDLCIdentifier = 0x42
const BIUDiagnosticResponseCode SDLCIdentifier = 0x43
const SDLCDiagnosticResponseCode SDLCIdentifier = 0x44
const SIUDiagnosticResponseCode SDLCIdentifier = 0x45
const DynamicStatusResponseCode SDLCIdentifier = 0x46
const AcknowledgeResponseCode SDLCIdentifier = 0x4F
const StaticStatusRequestCode SDLCIdentifier = 0x10
const SendDetectDataCode SDLCIdentifier = 0x11
const ConfigBIURequestCode SDLCIdentifier = 0x12
const BIUDiagnosticRequestCode SDLCIdentifier = 0x13
const SDLCDiagnosticRequestCode SDLCIdentifier = 0x14
const DynamicStatusRequestCode SDLCIdentifier = 0x15
const SIUDiagnosticRequestCode SDLCIdentifier = 0x16

func (id SDLCIdentifier) String() string {
	switch id {
	case StaticStatusResponseCode:
		return "StaticStatusResponseCode"
	case CMUFrameStreamCode:
		return "CMUFrameStreamCode"
	case DateTimeStreamCode:
		return "DateTimeStreamCode"
	case BIUDiagnosticResponseCode:
		return "BIUDiagnosticResponseCode"
	case SDLCDiagnosticResponseCode:
		return "SDLCDiagnosticResponseCode"
	case SIUDiagnosticResponseCode:
		return "SIUDiagnosticResponseCode"
	case DynamicStatusResponseCode:
		return "DynamicStatusResponseCode"
	case AcknowledgeResponseCode:
		return "AcknowledgeResponseCode"
	case StaticStatusRequestCode:
		return "StaticStatusRequestCode"
	case SendDetectDataCode:
		return "SendDetectDataCode"
	case ConfigBIURequestCode:
		return "ConfigBIURequestCode"
	case BIUDiagnosticRequestCode:
		return "BIUDiagnosticRequestCode"
	case SDLCDiagnosticRequestCode:
		return "SDLCDiagnosticRequestCode"
	case DynamicStatusRequestCode:
		return "DynamicStatusRequestCode"
	case SIUDiagnosticRequestCode:
		return "SIUDiagnosticRequestCode"
	default:
		return "Unknown"

	}
}
