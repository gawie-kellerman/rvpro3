package joystick

import (
	"fmt"
	"time"

	"github.com/warthog618/go-gpiocdev"
	"rvpro3/radarvision.com/internal/devices/lcd/interfaces"
	"rvpro3/radarvision.com/internal/hardware/m48"
	"rvpro3/radarvision.com/utils"
	"rvpro3/radarvision.com/utils/device/gpio"
)

const JoystickServiceName = "Joystick.Service"

type JoystickService struct {
	Metrics   JoystickServiceMetrics
	IsEnabled bool
	OpenErr   error
	Chips     gpio.Chips `json:"-"`
	LastRise  time.Time
	Debounce  utils.Milliseconds
	Consumer  interfaces.IPageManager
}

type JoystickServiceMetrics struct {
	ErrPortOpenCount     *utils.Metric
	ErrListenCount       *utils.Metric
	ErrUnmappedPortCount *utils.Metric
	ErrNoConsumerCount   *utils.Metric
	PortOpenCount        *utils.Metric
	JoystickEventCount   *utils.Metric
	utils.MetricsInitMixin
}

func (j *JoystickService) GetServiceName() string {
	return JoystickServiceName
}

func (j *JoystickService) Init() {
	j.Metrics.InitMetrics(j.GetServiceName(), &j.Metrics)
	j.Chips.Init()
}

func (j *JoystickService) InitFromSettings(settings *utils.Settings) {
	j.IsEnabled = settings.Basic.GetBool("joystick.enabled", true)
	j.Debounce = settings.Basic.GetMilliseconds("joystick.debounce", 200)
}

func (j *JoystickService) Start(state *utils.State, settings *utils.Settings) {
	if !j.IsEnabled {
		return
	}

	j.Init()
	state.Set(j.GetServiceName(), j)

	j.registerPort(int(m48.RightPort))
	j.registerPort(int(m48.LeftPort))
	j.registerPort(int(m48.UpPort))
	j.registerPort(int(m48.DownPort))
	j.registerPort(int(m48.EnterPort))
	j.registerPort(int(m48.EscapePort))

	if j.OpenErr != nil {
		fmt.Println(j.OpenErr)
	}

}

func (j *JoystickService) registerPort(portNo int) {
	if j.OpenErr != nil {
		return
	}

	var chip *gpio.Chip
	chip, j.OpenErr = j.Chips.OpenByPort(portNo)

	if j.OpenErr != nil {
		j.Metrics.ErrPortOpenCount.Inc(1)
	}

	_, j.OpenErr = chip.ListenToLine(gpio.Util.GetOffset(portNo))
	if j.OpenErr != nil {
		j.Metrics.ErrListenCount.Inc(1)
	}

	chip.OnHandleGPIO = j.onGPIOCallback

}

func (j *JoystickService) onGPIOCallback(chip *gpio.Chip, event gpiocdev.LineEvent) {
	portNo := m48.JoystickPort(gpio.Util.GetPortNo(chip.No, event.Offset))
	now := time.Now()

	switch event.Type {
	case gpiocdev.LineEventRisingEdge:
		if !j.Debounce.Expired(now, j.LastRise) {
			return
		}

	case gpiocdev.LineEventFallingEdge:
		return
	}

	j.LastRise = time.Now()

	switch portNo {
	case m48.LeftPort:
		j.sendEvent(interfaces.LeftPressed)
	case m48.RightPort:
		j.sendEvent(interfaces.RightPressed)
	case m48.UpPort:
		j.sendEvent(interfaces.UpPressed)
	case m48.DownPort:
		j.sendEvent(interfaces.DownPressed)
	case m48.EscapePort:
		j.sendEvent(interfaces.EscapePressed)
	case m48.EnterPort:
		j.sendEvent(interfaces.EnterPressed)
	default:
		j.Metrics.ErrUnmappedPortCount.Inc(1)
	}
}

func (j *JoystickService) IsDebounce(now time.Time) bool {
	return !j.Debounce.Expired(now, j.LastRise)
}

func (j *JoystickService) sendEvent(event interfaces.JoystickEvent) {
	if j.Consumer == nil {
		j.Consumer = utils.GlobalState.Get("LCD.Page.Service").(interfaces.IPageManager)
	}

	if j.Consumer == nil {
		j.Metrics.ErrNoConsumerCount.Inc(1)
		return
	}

	j.Consumer.OnJoystick(event)
	j.Metrics.JoystickEventCount.Inc(1)
}
