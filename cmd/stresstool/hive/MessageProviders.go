package hive

import (
	"rvpro3/radarvision.com/cmd/stresstool/config"
)

type MessageProviders struct {
	providers []*MessageProvider
	Latches   []*Latch
}

func (p *MessageProviders) Init(types []config.MessageType) {
	p.Latches = make([]*Latch, 0, 10)
	p.providers = make([]*MessageProvider, len(types))

	for i, t := range types {
		provider := new(MessageProvider)
		p.providers[i] = provider
		p.providers[i].Init(t.Name, t.Directory, t.LatchNo, t.LatchDurationSecs)

		latch := p.latch(t.LatchNo)

		if latch == nil {
			latch = &Latch{
				LatchNo:   t.LatchNo,
				Index:     -1,
				Providers: make([]*MessageProvider, 0, 10),
			}
			p.Latches = append(p.Latches, latch)
		}

		latch.Providers = append(latch.Providers, provider)
	}
}

func (p *MessageProviders) latch(latchNo int) *Latch {
	// Latch 0 is always a new latch
	if latchNo == 0 {
		return nil
	}

	for _, l := range p.Latches {
		if l.LatchNo == latchNo {
			return l
		}
	}
	return nil
}
