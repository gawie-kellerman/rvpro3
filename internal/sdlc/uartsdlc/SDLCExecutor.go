package uartsdlc

import (
	"time"

	"github.com/rs/zerolog/log"
	"rvpro3/radarvision.com/utils"
)

const SDLCStaticStatusStateName = "SDLC.StaticStatus"
const SDLCExecutorServiceStateName = "SDLC.Executor.Service"
const SDLCExecutorMetricsAt = "SDLC.Executor"

//type sdlcExecutorSettings[T SDLCExecutorService] struct{}
//
//var SDLCExecutorSettings = sdlcExecutorSettings{}

const sdlcUARTStaticStatusRequestEvery = "SDLC.UART.StaticRequest.Every"

type sdlcExecutorSettings struct{}

type SDLCExecutorService struct {
	Metronome             utils.Metronome
	Terminate             bool
	Terminated            bool
	sdlcService           *SDLCService
	Now                   time.Time
	StaticRequestOn       time.Time
	StaticRequestInterval time.Duration
	StaticStatus          *StaticStatus

	DecodeErrCount           *utils.Metric
	DecodeErrBytes           *utils.Metric
	StaticStatusRequests     *utils.Metric
	StaticStatusResponses    *utils.Metric
	StaticStatusRequestEvery time.Duration
}

func (s *SDLCExecutorService) SetupDefaults(config *utils.Settings) {
	config.SetSettingAsMillis(sdlcUARTStaticStatusRequestEvery, 5000)
}

func (s *SDLCExecutorService) SetupAndStart(state *utils.State, config *utils.Settings) {
	// Don't start the service if the SDLC UART is not enabled
	if !config.GetSettingAsBool(sdlcUARTEnabled) {
		return
	}

	service := new(SDLCExecutorService)
	service.InitFromConfig(config)
	service.Start()

	state.Set(SDLCExecutorServiceStateName, service)
}

func (s *SDLCExecutorService) InitFromConfig(config *utils.Settings) {
	s.StaticStatusRequestEvery = config.GetSettingAsMillis(sdlcUARTStaticStatusRequestEvery)
}

func (s *SDLCExecutorService) GetServiceName() string {
	return SDLCExecutorServiceStateName
}

func (s *SDLCExecutorService) GetServiceNames() []string {
	return nil
}

func (s *SDLCExecutorService) InitMetrics() {
	gm := &utils.GlobalMetrics
	s.StaticStatusRequests = gm.U64(SDLCExecutorMetricsAt, "Static Status Requests")
	s.StaticStatusResponses = gm.U64(SDLCExecutorMetricsAt, "Static Status Responses")
}

func (s *SDLCExecutorService) Start() {
	s.InitMetrics()
	utils.GlobalState.Set(SDLCExecutorServiceStateName, s)
	s.Terminated = false
	s.Terminate = false
	s.Metronome.CycleDuration = 100 * time.Millisecond
	s.Metronome.IsReal = false
	s.sdlcService = utils.GlobalState.Get(SDLCServiceName).(*SDLCService)
	s.StaticRequestInterval = time.Duration(10) * time.Second

	if s.sdlcService == nil {
		panic("SDLC s is not running")
	}

	s.sdlcService.OnReadMessage = s.OnReadMessage
	s.StaticStatus = utils.GlobalState.Set(SDLCStaticStatusStateName, new(StaticStatus)).(*StaticStatus)

	go s.execute()
}

func (s *SDLCExecutorService) execute() {
	s.Metronome.Start()

	for !s.Terminate {
		s.Now = time.Now()
		s.doStaticStatusRequest()

		s.Metronome.AwaitClick()
	}
	s.Terminated = true
}

func (s *SDLCExecutorService) doStaticStatusRequest() {
	if !utils.Time.IsExpired(
		s.Now,
		s.StaticRequestOn,
		s.StaticRequestInterval,
	) {
		return
	}

	s.StaticStatusRequests.Add(1, s.Now)
	encoder := SDLCRequestEncoder{}
	data, err := encoder.StaticStatus()

	if err != nil {
		log.Err(err).Msg("SDLCExecutorService.doStaticStatusRequest")
	}

	s.sdlcService.Write(data)
	s.StaticRequestOn = s.Now
}

func (s *SDLCExecutorService) OnReadMessage(_ *SDLCService, data []byte) {
	decoder := SDLCResponseDecoder{}

	if err := decoder.Init(data); err != nil {
		s.DecodeErrCount.Add(1, s.Now)
		s.DecodeErrBytes.Add(len(data), s.Now)
		return
	}

	switch decoder.GetIdentifier() {
	case StaticStatusResponseCode:
		s.onStaticResponse(&decoder)
	}
}

func (s *SDLCExecutorService) onStaticResponse(decoder *SDLCResponseDecoder) {
	status, err := decoder.GetStaticStatus()
	if err != nil {
		// TODO: Add Metric
		log.Err(err).Msg("SDLCExecutorService.onStaticResponse")
		return
	}

	s.StaticStatusResponses.Add(1, s.Now)

	// TODO: Processing surrounding BUI needed

	*s.StaticStatus = status
}
