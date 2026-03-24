package pages

import (
	"strings"

	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
	"rvpro3/radarvision.com/internal/services/ping"
	"rvpro3/radarvision.com/utils"
)

type LcdHomePage struct {
	text      string
	pingStats *ping.PingStats
	LcdMixinPage
}

func (l *LcdHomePage) OnJoystick(
	press interfaces.JoystickEvent,
) bool {
	l.text = press.String()
	l.RedrawNeeded = true

	switch press {
	case interfaces.RightPressed:
		l.Manager.ShowPage(&LcdAboutPage{})
	default:
		break
	}

	return true
}

func (l *LcdHomePage) Refresh(canvas interfaces.ILcdCanvas) {
	canvas.ClearScreen()

	l.DrawHeader(canvas)

	//canvas.DrawStrLn("Pressed: " + l.text)
	//
	//radarStatuses, cameraStatuses := l.getPingStatuses()
	//canvas.DrawStrLn("Radar:" + radarStatuses)
	//canvas.DrawStrLn("Camera:" + cameraStatuses)

	l.LastUpdateOn = utils.Time.Exact()
}

func (l *LcdHomePage) IsRedrawNeeded() bool {
	return l.GetRedrawByTime(utils.Time.Approx())
}

func (l *LcdHomePage) getPingStatuses() (string, string) {
	var ok bool

	camStr := strings.Builder{}
	radStr := strings.Builder{}

	if l.pingStats == nil {
		l.pingStats, ok = utils.GlobalState.Get(ping.PingStatsStateName).(*ping.PingStats)

		if !ok {
			return "No Pings", "No Pings"
		}
	}

	for _, stat := range l.pingStats.List {
		switch stat.DeviceType {

		case ping.DeviceTypeCamera:
			stat.AddStatusTo(&camStr)

		default:
			stat.AddStatusTo(&radStr)
		}
	}
	return radStr.String(), camStr.String()
}
