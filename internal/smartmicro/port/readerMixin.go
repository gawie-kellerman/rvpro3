package port

import "encoding/binary"

type readerMixin struct {
	Buffer       []byte
	StartOffset  int
	Order        binary.ByteOrder
	VersionMajor int
	VersionMinor int
	DetailLen    int
}

func (r *readerMixin) initBuffer(buffer []byte) {
	r.Buffer = buffer

	th := TransportHeaderReader{
		Buffer: buffer,
	}

	ph := PortHeaderReader{
		Buffer:      buffer,
		StartOffset: int(th.GetHeaderLength()),
	}

	r.StartOffset = int(th.GetHeaderLength()) + ph.GetHeaderLength()
	r.VersionMajor = int(ph.GetPortMajorVersion())
	r.VersionMinor = int(ph.GetPortMinorVersion())
	r.Order = ph.GetBodyOrder().ToGo()
}

// InitTransport likely did TransportHeader Validation already.
// TODO: Parameterize the validation check
func (r *readerMixin) InitTransport(th *TransportHeaderReader) error {
	th.Buffer = r.Buffer
	if err := th.CheckFormat(); err != nil {
		return err
	}

	if err := th.CheckCRC(); err != nil {
		return err
	}

	return nil
}

// InitPort likely did PortHeader validation already.
// TODO: Parameterize the validation check
func (r *readerMixin) InitPort(ph *PortHeaderReader) error {
	th := TransportHeaderReader{}

	if err := r.InitTransport(&th); err != nil {
		return err
	}

	ph.Buffer = r.Buffer
	ph.StartOffset = int(th.GetHeaderLength())
	return ph.Check()
}
