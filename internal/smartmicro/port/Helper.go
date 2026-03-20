package port

type helper struct{}

var Helper helper

func (helper) GetHeaders(bytes []byte) (TransportHeaderReader, PortHeaderReader) {
	th := TransportHeaderReader{
		Buffer: bytes,
	}

	ph := PortHeaderReader{
		Buffer:      bytes,
		StartOffset: int(th.GetHeaderLength()),
	}
	return th, ph
}
