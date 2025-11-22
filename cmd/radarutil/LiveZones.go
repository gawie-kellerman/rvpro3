package main

import "sync"

type LiveConfig struct {
	DistanceUnit string             `json:"distance_unit,omitempty"`
	Radars       []*LiveConfigRadar `json:"radars"`
	Assignments  int                `json:"-"`
	mutex        sync.Mutex
}

func (l *LiveConfig) IsSegmentComplete(radarIP string) bool {
	radar := l.radar(radarIP)

	for _, zone := range radar.Zones {
		if !zone.IsSegmentComplete() {
			return false
		}
	}
	return true
}

func (l *LiveConfig) CalcAssignmentsNeeded() int {
	res := 0

	for _, radar := range l.Radars {
		res += radar.CalcAssignmentsNeed()
	}

	return res
}

func (l *LiveConfig) IsComplete() bool {
	return l.Assignments == l.CalcAssignmentsNeeded() && l.Assignments > 0
}

func (l *LiveConfig) SetupZones(radarIP string, nofZones int) {
	radar := l.radar(radarIP)
	radar.Zones = make([]*LiveConfigZone, nofZones)

	for i, _ := range radar.Zones {
		radar.Zones[i] = &LiveConfigZone{}
		radar.Zones[i].ZoneNo = i + 1
		radar.Zones[i].Description = radarIP
	}
}

func (l *LiveConfig) ZoneAndCoordBySegment(radarIP string, segmentNo int) (int, int) {
	radar := l.radar(radarIP)

	for i, zone := range radar.Zones {
		if segmentNo >= len(zone.Coords) {
			segmentNo -= len(zone.Coords)
		} else {
			return i, segmentNo
		}
	}
	panic("unreachable")
}

func (l *LiveConfig) SetupSegments(radarIP string, zoneNo int, nofSegments int) {
	radar := l.radar(radarIP)
	zone := radar.Zones[zoneNo]
	zone.Coords = make([]LiveConfigCoord, nofSegments)
}

func (l *LiveConfig) SetWidth(radarIP string, zoneNo int, width float32) {
	l.Assignments++
	radar := l.radar(radarIP)
	zone := radar.Zones[zoneNo]
	zone.Width = width
}

func (l *LiveConfig) SetTrigger(radarIP string, zoneNo int, trigger int) {
	l.Assignments++
	radar := l.radar(radarIP)
	zone := radar.Zones[zoneNo]
	zone.Trigger = trigger
}

func (l *LiveConfig) SetX(radarIP string, zoneNo int, coordNo int, x float32) {
	l.Assignments++
	radar := l.radar(radarIP)
	zone := radar.Zones[zoneNo]
	zone.Coords[coordNo].X = x
}

func (l *LiveConfig) SetY(radarIP string, zoneNo int, coordNo int, y float32) {
	l.Assignments++
	radar := l.radar(radarIP)
	zone := radar.Zones[zoneNo]
	zone.Coords[coordNo].Y = y
}

// radar assumes, from implementation of the instruction service,
// the there is currency on the radar array, but not on any of its children,
// hence the reason for only protected this private radar method
func (l *LiveConfig) radar(radarIP string) *LiveConfigRadar {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	for _, radar := range l.Radars {
		if radar.RadarIP == radarIP {
			return radar
		}
	}
	radar := &LiveConfigRadar{}
	radar.RadarIP = radarIP
	l.Radars = append(l.Radars, radar)
	return radar
}

func (l *LiveConfig) Init() {
	l.DistanceUnit = "m"
	l.Radars = make([]*LiveConfigRadar, 0, 4)
}

func (l *LiveConfig) GetSegmentCount(radarIP string) int {
	res := 0

	radar := l.radar(radarIP)
	for _, zone := range radar.Zones {
		res += len(zone.Coords)
	}

	return res
}

type LiveConfigRadar struct {
	RadarIP string            `json:"radar_ip"`
	Zones   []*LiveConfigZone `json:"zones"`
}

func (r *LiveConfigRadar) CalcAssignmentsNeed() int {
	res := 0

	for _, zone := range r.Zones {
		res += zone.CalcAssignmentsNeeded()
	}

	return res
}

type LiveConfigZone struct {
	ZoneNo      int               `json:"zone_no"`
	Description string            `json:"description"`
	Coords      []LiveConfigCoord `json:"coords"`
	Trigger     int               `json:"trigger"`
	Width       float32           `json:"width"`
}

// CalcAssignmentsNeeded calculates the total number of assignments needed
// to complete the zone configuration
func (z *LiveConfigZone) CalcAssignmentsNeeded() int {
	return len(z.Coords)*2 + 2
}

func (z *LiveConfigZone) IsSegmentComplete() bool {
	return z.Coords != nil
}

type LiveConfigCoord struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}
