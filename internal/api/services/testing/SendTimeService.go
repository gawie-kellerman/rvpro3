package testing

import (
	"time"

	"rvpro3/radarvision.com/internal/api/services/web"
	"rvpro3/radarvision.com/utils"
)

type SendTimeService struct {
	web       *web.WebService
	Terminate bool
}

func (s *SendTimeService) SetupDefaults(settings *utils.Settings) {
	settings.SetSettingAsBool("http.socket.timer.enabled", true)
}

func (s *SendTimeService) SetupAndStart(state *utils.State, config *utils.Settings) {
	if utils.GlobalSettings.GetSettingAsBool("http.socket.timer.enabled") {
		s.Terminate = false
		s.web = utils.GlobalState.Get(web.WebServiceName).(*web.WebService)
		if s.web != nil {
			go s.run()
		}
	}
}

func (s *SendTimeService) GetServiceName() string {
	return "Send.Time.Service"
}

func (s *SendTimeService) GetServiceNames() []string {
	return nil
}

func (s *SendTimeService) run() {
	for !s.Terminate {
		if s.web.IsAnySubscribed(web.SocketTime) {
			now := time.Now()
			msg := web.SocketMessage{}

			msg.Init()
			msg.SetType("time-stream")
			msg.Set("Value", now.Format(utils.DisplayDateTimeMS))
			msg.Set("Location", now.Location().String())
			zone, offset := now.Zone()
			msg.Set("Zone", zone)
			msg.SetInt("Offset", offset)
			s.web.Broadcast(msg.ToPayload(web.SocketTime))
		}

		time.Sleep(1 * time.Second)
	}
}
