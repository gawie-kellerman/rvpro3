package port

import "github.com/pkg/errors"

// Protocol is a zero base struct
type Protocol struct{}
type ProtocolType uint8

const (
	PtSmartMicroCANv1 ProtocolType = iota
	PtATXMega
	PtSTG
	PtUnused
	PtSmartMicroCANv2
	PtDebugData
	PtLogMessageData
	PtAliveProtocol
	PtSmartMicroPort

	// PtSmartMicroIAP is Interview Application Protocol
	PtSmartMicroIAP
)

var ErrPayloadCRC = errors.New("payload crc16 error")

func (p ProtocolType) ToString() string {
	switch p {
	case PtSmartMicroCANv1:
		return "smartmicro can v1"

	case PtATXMega:
		return "atx mega"

	case PtSTG:
		return "stg"

	case PtSmartMicroCANv2:
		return "smartmicro can v2"

	case PtDebugData:
		return "debug data"

	case PtLogMessageData:
		return "log message data"

	case PtAliveProtocol:
		return "alive protocol"

	case PtSmartMicroPort:
		return "smart micro port"

	default:
		return "unknown"
	}
}
