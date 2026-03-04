package testing

import (
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/internal/api/services/web"
	"rvpro3/radarvision.com/internal/general"
	"rvpro3/radarvision.com/utils"
)

type SendTimeSocketService struct {
	web       *web.WebService
	IsEnabled bool
	Terminate bool
}

func (s *SendTimeSocketService) InitFromSettings(settings *utils.Settings) {
	s.IsEnabled = settings.Basic.GetBool("feature.http.timersocket.enabled", true)
}

func (s *SendTimeSocketService) Start(state *utils.State, settings *utils.Settings) {
	if !general.ServiceHelper.ShouldStart(state, settings, s) {
		return
	}

	if !s.IsEnabled {
		return
	}

	// No metrics to initialize

	s.Terminate = false
	s.web = utils.GlobalState.Get(web.WebServiceName).(*web.WebService)
	if s.web != nil {
		go s.run()
	} else {
		s.IsEnabled = false
		log.Warn().Msg("Unable to start time socket as WebService is not enabled")
	}
}

func (s *SendTimeSocketService) GetServiceName() string {
	return "Send.Time.Socket.Service"
}

func (s *SendTimeSocketService) run() {
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
