package utils

import (
	"time"

	"github.com/rs/zerolog/log"
)

type ErrorLoggerMixin struct {
	lastError       error
	lastErrorCount  int
	lastErrorTime   time.Time
	LogRepeatMillis time.Duration
}

func (el *ErrorLoggerMixin) LogError(msg string, err error) {
	now := time.Now()

	if el.lastError == nil {
		if err == nil {
			el.lastErrorCount = 0
			return
		}

		el.lastErrorCount = 0
		el.lastError = err
		el.lastErrorTime = now
		log.Error().Err(err).Msg(msg)
		return
	}
	if err == nil {
		el.lastErrorCount = 0
		el.lastError = err
		return
	}

	if el.lastError.Error() != err.Error() {
		if el.lastErrorCount > 1 {
			log.Err(el.lastError).
				Str("since", el.lastErrorTime.Format(DisplayTimeMS)).
				Int("repeat", el.lastErrorCount).
				Msg(msg)
		}

		log.Err(err).Msg(msg)

		el.lastErrorTime = now
		el.lastErrorCount = 0
		el.lastError = err
	} else {
		if now.Sub(el.lastErrorTime) > el.LogRepeatMillis {
			if el.lastErrorCount > 0 {
				log.Err(el.lastError).
					Str("since", el.lastErrorTime.Format(DisplayTimeMS)).
					Int("repeat", el.lastErrorCount).
					Msg(msg)
				el.lastErrorTime = now
				el.lastErrorCount = 0
			}
		} else {
			// Same error
			el.lastErrorCount += 1
		}
	}
}

func (el *ErrorLoggerMixin) ClearError() {
	el.lastError = nil
	el.lastErrorCount = 0
}
