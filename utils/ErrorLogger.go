package utils

import (
	"time"

	"github.com/rs/zerolog/log"
)

type ErrorLoggerMixin struct {
	LastError       error
	LastErrorCount  int
	LastErrorTime   time.Time
	LogRepeatMillis time.Duration
}

func (el *ErrorLoggerMixin) LogError(msg string, err error) {
	now := time.Now()

	if el.LastError == nil {
		if err == nil {
			el.LastErrorCount = 0
			return
		}

		el.LastErrorCount = 0
		el.LastError = err
		el.LastErrorTime = now
		log.Error().Err(err).Msg(msg)
		return
	}
	if err == nil {
		el.LastErrorCount = 0
		el.LastError = err
		return
	}

	if el.LastError.Error() != err.Error() {
		if el.LastErrorCount > 1 {
			log.Err(el.LastError).
				Str("since", el.LastErrorTime.Format(DisplayTimeMS)).
				Int("repeat", el.LastErrorCount).
				Msg(msg)
		}

		log.Err(err).Msg(msg)

		el.LastErrorTime = now
		el.LastErrorCount = 0
		el.LastError = err
	} else {
		if now.Sub(el.LastErrorTime) > el.LogRepeatMillis {
			if el.LastErrorCount > 0 {
				log.Err(el.LastError).
					Str("since", el.LastErrorTime.Format(DisplayTimeMS)).
					Int("repeat", el.LastErrorCount).
					Msg(msg)
				el.LastErrorTime = now
				el.LastErrorCount = 0
			}
		} else {
			// Same error
			el.LastErrorCount += 1
		}
	}
}

func (el *ErrorLoggerMixin) ClearError() {
	el.LastError = nil
	el.LastErrorCount = 0
}
