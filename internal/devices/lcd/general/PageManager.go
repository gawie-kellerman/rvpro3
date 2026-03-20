package general

import (
	"bytes"

	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
	"rvpro3/radarvision.com/utils"
)

const lcdWidth = 128
const lcdHeight = 64

type PageManager struct {
	HomePage    interfaces.IPage
	CurrentPage interfaces.IPage
	Metrics     PageManagerMetrics

	CurrentMono LcdCanvas
	CompareMono LcdCanvas
}

type PageManagerMetrics struct {
	NilPageRefreshCount *utils.Metric
	HardRefreshCount    *utils.Metric
	SoftRefreshCount    *utils.Metric
	PageDrawCount       *utils.Metric
	PageSkipCount       *utils.Metric
}

func (m *PageManager) Init() {
	m.CurrentMono.Init(lcdWidth, lcdHeight)
	m.CompareMono.Init(lcdWidth, lcdHeight)
}

func (m *PageManager) ShowPage(page interfaces.IPage) {
	m.CurrentPage = page
}

func (m *PageManager) GetPage() interfaces.IPage {
	return m.CurrentPage
}

func (m *PageManager) GetHomePage() interfaces.IPage {
	return m.HomePage
}

func (m *PageManager) SetHomePage(homePage interfaces.IPage) {
	m.HomePage = homePage
}

func (m *PageManager) OnRedraw(hardRefresh bool) {
	if m.CurrentPage == nil {
		m.Metrics.NilPageRefreshCount.Inc(1)
		return
	} else {
		if hardRefresh {
			m.Metrics.HardRefreshCount.Inc(1)
		} else {
			m.Metrics.SoftRefreshCount.Inc(1)
		}
	}

	m.CurrentPage.OnRefresh(hardRefresh)

	for pageNo := 0; pageNo < 8; pageNo++ {
		cur := m.CurrentMono.GetPage(pageNo)
		cmp := m.CompareMono.GetPage(pageNo)

		if bytes.Compare(cmp, cur) != 0 {
			m.Metrics.PageDrawCount.Inc(1)
			m.CompareMono.Copy(cur, m.CurrentMono.GetPageStart(pageNo))
		} else {
			m.Metrics.PageSkipCount.Inc(1)
		}
	}
}

func (m *PageManager) OnJoystick(event interfaces.JoystickEvent) {
	if m.CurrentPage == nil {
		m.Metrics.NilPageRefreshCount.Inc(1)
		return
	}

	if m.CurrentPage.OnJoystick(event) {
		m.ShowPage(m.HomePage)
	}
}
