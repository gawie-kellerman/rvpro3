package service

import (
	"time"

	"rvpro3/radarvision.com/utils"
)

type MixinDataService struct {
	now               time.Time
	Terminating       bool
	Terminated        bool
	LoopGuard         utils.LoopGuard
	RetryGuard        utils.RetryGuard
	OnError           func(sender any, err error)
	OnOpenConnection  func(sender any)
	OnCloseConnection func(sender any)
	OnStart           func(sender any)
	OnTerminate       func(sender any)
	OnLoop            func(sender any)
	OnNoData          func(sender any)
}

func (s *MixinDataService) OnErrorCallback(sender any, err error) {
	if s.OnError != nil {
		s.OnError(sender, err)
	}
}

func (s *MixinDataService) OnConnectCallback(sender any) {
	if s.OnOpenConnection != nil {
		s.OnOpenConnection(sender)
	}
}

func (s *MixinDataService) OnDisconnectCallback(sender any) {
	if s.OnCloseConnection != nil {
		s.OnCloseConnection(sender)
	}
}

func (s *MixinDataService) OnStartCallback(sender any) {
	if s.OnStart != nil {
		s.OnStart(sender)
	}
}

func (s *MixinDataService) OnTerminateCallback(sender any) {
	if s.OnTerminate != nil {
		s.OnTerminate(sender)
	}
}

func (s *MixinDataService) onLoopCallback(sender any) {
	if s.OnLoop != nil {
		s.OnLoop(sender)
	}
}

func (s *MixinDataService) onNoDataCallback(sender any) {
	if s.OnNoData != nil {
		s.OnNoData(sender)
	}
}
