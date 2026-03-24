package ping

import "rvpro3/radarvision.com/utils"

const PingStatsStateName = "Ping.Stats"

type PingStats struct {
	Index int
	List  []*PingStat
}

func (p *PingStats) Init() {
	p.Index = 0
	p.List = make([]*PingStat, 0, 8)
	utils.GlobalState.Set(PingStatsStateName, p)
}

func (p *PingStats) Add(ip4 utils.IP4, deviceType string) *PingStat {
	res := p.Find(ip4)

	if res != nil {
		return res
	}

	res = &PingStat{}
	res.Init(ip4, deviceType)
	p.List = append(p.List, res)
	return res
}

func (p *PingStats) Find(ip4 utils.IP4) *PingStat {
	ip4Str := ip4.ToIPString()

	for _, stat := range p.List {
		if stat.DriverStat.Addr == ip4Str {
			return stat
		}
	}

	return nil
}

func (p *PingStats) GetNext() (res *PingStat, isWrapped bool) {
	if len(p.List) == 0 {
		return nil, true
	}

	if p.Index >= len(p.List) {
		p.Index = 0
		isWrapped = true
	}

	res = p.List[p.Index]
	p.Index++
	return res, isWrapped
}
