package pages

import (
	"strconv"
	"time"

	"rvpro3/radarvision.com/internal/constants"
	"rvpro3/radarvision.com/internal/devices/lcd/fonts"
	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
	"rvpro3/radarvision.com/internal/router/server"
	"rvpro3/radarvision.com/utils"
)

var connectIcon = []uint16{55, 54, 53}

type LcdMixinPage struct {
	Manager      interfaces.IPageManager
	LastUpdateOn time.Time
	RedrawNeeded bool
	appInfo      *utils.AppInfo
	router       *server.RouterServerService
}

func (l *LcdMixinPage) BeforeShow(manager interfaces.IPageManager) {
	var ok bool
	l.Manager = manager
	l.RedrawNeeded = true

	l.appInfo, ok = utils.GlobalState.Get(constants.AppInfoStateName).(*utils.AppInfo)

	if !ok {
		l.appInfo = &utils.AppInfo{}
	}
}

func (l *LcdMixinPage) GetRedrawByTime(now time.Time) bool {
	return l.RedrawNeeded || now.Sub(l.LastUpdateOn).Seconds() > 1
}

func (l *LcdMixinPage) DrawHeader(canvas interfaces.ILcdCanvas) {
	// Draw App Version
	canvas.SetXY(0, 0)
	canvas.DrawStr("RVC ")
	canvas.DrawStr(l.appInfo.Version)
	canvas.SetX(42)

	// Draw Router Connectivity
	clients := l.getRouterClients()
	switch clients {
	case -1:
		canvas.DrawSymbol(fonts.WingdingsA5D1, fonts.IconBrokenConnection)
	case 0:
		canvas.DrawSymbol(fonts.WingdingsA5D1, fonts.IconDisconnected)

	default:
		canvas.DrawSymbol(fonts.WingdingsA5D1, fonts.IconConnected)
		canvas.MoveCursorBy(2, 0)
		canvas.DrawStr(strconv.Itoa(clients))
	}

	// Draw Time (top right)
	canvas.DrawRight(canvas.Width(), time.Now().Format(time.TimeOnly))
	canvas.DrawLn()

	canvas.HorzLine(canvas.GetY() + 1)
	canvas.MoveCursorBy(0, 2)
}

func (l *LcdMixinPage) getRouterClients() int {
	if l.router == nil {
		var ok bool
		l.router, ok = utils.GlobalState.Get(constants.RouterServerService).(*server.RouterServerService)

		if !ok {
			return -1
		}
	}

	return len(l.router.Server.Connections)
}
