package general

import (
	"bytes"
	"time"

	"rvpro3/radarvision.com/internal/devices/lcd/fonts"
	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
	"rvpro3/radarvision.com/utils"
)

const lcdWidth = 128
const lcdHeight = 64

const LcdPageServiceName = "LCD.Page.Service"

type LcdPageService struct {
	Metrics         PageManagerMetrics
	DeviceName      string
	DeviceAddress   int
	IsEnabled       bool
	OpenErr         error
	Driver          LcdDriver           `json:"-"`
	HomePage        interfaces.ILcdPage `json:"-"`
	CurrentPage     interfaces.ILcdPage `json:"-"`
	ScreenSaverPage interfaces.ILcdPage `json:"-"`
	CurCanvas       LcdCanvas           `json:"-"`
	BakCanvas       LcdCanvas           `json:"-"`
	Terminate       bool
	Terminated      bool
	RefreshCooldown utils.Milliseconds
	LastKeypress    time.Time
	ScreenSaverMins int
}

type PageManagerMetrics struct {
	ErrDeviceOpenCount  *utils.Metric
	ErrDrawPageCount    *utils.Metric
	DeviceOpenCount     *utils.Metric
	NilPageRefreshCount *utils.Metric
	RefreshCount        *utils.Metric
	PageDrawCount       *utils.Metric
	PageSkipCount       *utils.Metric
	JoystickPressCount  *utils.Metric
	utils.MetricsInitMixin
}

func (m *LcdPageService) InitBeforeStart() {
	m.Metrics.InitMetrics(m.GetServiceName(), &m.Metrics)
	m.CurCanvas.Init(lcdWidth, lcdHeight, fonts.NotoSansA6D2)
	m.BakCanvas.Init(lcdWidth, lcdHeight, fonts.NotoSansA6D2)
	m.BakCanvas.Fill(0b01010101)
}

func (m *LcdPageService) GetServiceName() string {
	return LcdPageServiceName
}

func (m *LcdPageService) InitFromSettings(settings *utils.Settings) {
	m.IsEnabled = settings.Basic.GetBool("lcd.enabled", true)
	m.DeviceName = settings.Basic.Get("lcd.device.name", "/dev/i2c-1")
	m.DeviceAddress = settings.Basic.GetInt("lcd.device.address", 0x3c)
	m.RefreshCooldown = settings.Basic.GetMilliseconds("lcd.refresh.cooldown", 100)
	m.ScreenSaverMins = settings.Basic.GetInt("lcd.screensaver.minutes", 1)
	m.LastKeypress = time.Now()
}

func (m *LcdPageService) Start(state *utils.State, _ *utils.Settings) {
	m.InitBeforeStart()
	state.Set(m.GetServiceName(), m)

	if !m.IsEnabled {
		return
	}

	m.OpenErr = m.Driver.Open(m.DeviceName, m.DeviceAddress)

	if m.OpenErr != nil {
		m.IsEnabled = false
		m.Metrics.ErrDeviceOpenCount.Inc(1)
	} else {
		m.Metrics.DeviceOpenCount.Inc(0)
	}

	m.InitBeforeStart()
	go m.run()
}

func (m *LcdPageService) run() {
	for !m.Terminate {
		cp := m.CurrentPage
		if cp != nil {
			if cp.IsRedrawNeeded() {
				m.redraw()
			}
		}

		now := utils.Time.Exact()

		if m.ScreenSaverPage != nil {
			if int(now.Sub(m.LastKeypress).Minutes()) > m.ScreenSaverMins {
				m.ShowPage(m.ScreenSaverPage)
			}
		}

		m.RefreshCooldown.Sleep()
	}
	m.Terminated = true
}

func (m *LcdPageService) ShowPage(page interfaces.ILcdPage) {
	if page != nil {
		page.BeforeShow(m)
		m.CurrentPage = page
	}
}

func (m *LcdPageService) GetPage() interfaces.ILcdPage {
	return m.CurrentPage
}

func (m *LcdPageService) GetHomePage() interfaces.ILcdPage {
	return m.HomePage
}

func (m *LcdPageService) SetHomePage(homePage interfaces.ILcdPage) {
	m.HomePage = homePage

	if m.CurrentPage == nil {
		m.ShowPage(homePage)
	}
}

func (m *LcdPageService) redraw() {
	if m.CurrentPage == nil {
		m.CurrentPage = m.HomePage
	}

	if m.CurrentPage == nil {
		m.Metrics.NilPageRefreshCount.Inc(1)
		return
	}

	m.Metrics.RefreshCount.Inc(1)

	m.CurrentPage.Refresh(&m.CurCanvas)

	for pageNo := 0; pageNo < 8; pageNo++ {
		cur := m.CurCanvas.GetPage(pageNo)
		cmp := m.BakCanvas.GetPage(pageNo)

		if bytes.Compare(cmp, cur) != 0 {
			m.Metrics.PageDrawCount.Inc(1)
			err := m.Driver.DrawPage(m.CurCanvas.GetPage(pageNo), pageNo)
			if err != nil {
				m.Metrics.ErrDrawPageCount.Inc(1)
			}

			m.BakCanvas.Copy(cur, m.CurCanvas.GetPageStart(pageNo))

		} else {
			m.Metrics.PageSkipCount.Inc(1)
		}
	}
}

func (m *LcdPageService) OnJoystick(event interfaces.JoystickEvent) {
	m.LastKeypress = utils.Time.Approx()

	if m.CurrentPage == nil {
		m.Metrics.NilPageRefreshCount.IncAt(1, utils.Time.Approx())
		return
	}

	m.Metrics.JoystickPressCount.IncAt(1, utils.Time.Approx())

	propagate := m.CurrentPage.OnJoystick(event)

	if event == interfaces.EscapePressed && propagate {
		m.ShowPage(m.HomePage)
	}
}
