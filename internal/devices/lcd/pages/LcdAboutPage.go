package pages

import (
	"time"

	"rvpro3/radarvision.com/internal/devices/lcd/fonts"
	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
)

type LcdAboutPage struct {
	lastUpdate     time.Time
	text           string
	isRedrawNeeded bool
}

func (l *LcdAboutPage) Init() {}

func (l *LcdAboutPage) OnJoystick(press interfaces.JoystickEvent) bool {
	return true
}

func (l *LcdAboutPage) Refresh(canvas interfaces.ILcdCanvas) {
	now := time.Now()

	if l.text == "" {
		l.text = "None"
	}

	canvas.ClearScreen()
	canvas.SetFont(fonts.NotoSansA6D2)
	canvas.SetXY(0, 0)

	canvas.DrawStrLn(now.String())
	l.lastUpdate = now
	l.isRedrawNeeded = false
}

func (l *LcdAboutPage) IsRedrawNeeded() bool {
	return l.isRedrawNeeded || time.Now().Sub(l.lastUpdate).Seconds() >= 1
}

func (l *LcdAboutPage) BeforeShow(manager interfaces.IPageManager) {
	l.isRedrawNeeded = true
}
