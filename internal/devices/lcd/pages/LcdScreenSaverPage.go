package pages

import (
	"strconv"

	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
	"rvpro3/radarvision.com/utils"
)

type LcdScreenSaverPage struct {
	LcdMixinPage
	Row int
}

func (l *LcdScreenSaverPage) OnJoystick(press interfaces.JoystickEvent) bool {
	l.Manager.ShowPage(l.Manager.GetHomePage())
	return true
}

func (l *LcdScreenSaverPage) Refresh(canvas interfaces.ILcdCanvas) {
	canvas.ClearScreen()
	canvas.SetXY(0, l.Row%20)
	canvas.DrawStrLn("Saver (" + strconv.Itoa(l.Row) + ")")
	canvas.DrawStrLn("Screen Saver")

	l.LastUpdateOn = utils.Time.Approx()
}

func (l *LcdScreenSaverPage) IsRedrawNeeded() bool {
	now := utils.Time.Approx()

	if int(now.Sub(l.LastUpdateOn).Minutes()) > 1 {
		l.Row += 1
		return true
	}
	return false
}
