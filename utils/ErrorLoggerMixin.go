package utils

import (
	"time"

	"github.com/rs/zerolog/log"
)

type ErrorLoggerMixin struct {
	Error          error
	Count          int
	Time           time.Time
	RepeatDuration Milliseconds
}

func (el *ErrorLoggerMixin) LogErrorAt(now time.Time, msg string, err error) {
	if el.Error == nil {
		if err == nil {
			return
		}

		el.Count = 0
		el.Error = err
		el.Time = now
		log.Error().Err(err).Msg(msg)
		return
	}

	if err == nil {
		el.Count = 0
		el.Error = err
		return
	}

	if el.Error.Error() != err.Error() {
		// Persist tail of previous error
		if el.Count > 1 {
			log.Err(el.Error).
				Str("since", el.Time.Format(DisplayTimeMS)).
				Int("repeat", el.Count).
				Msg(msg)
		}

		// Log new error
		log.Err(err).Msg(msg)
		el.Time = now
		el.Count = 0
		el.Error = err
	} else {
		if now.Sub(el.Time) > time.Duration(el.RepeatDuration) {
			if el.Count > 0 {
				log.Err(el.Error).
					Str("since", el.Time.Format(DisplayTimeMS)).
					Int("repeat", el.Count).
					Msg(msg)
				el.Time = now
				el.Count = 0
			}
		} else {
			// Same error
			el.Count += 1
		}
	}

}

func (el *ErrorLoggerMixin) Clear() {
	el.Error = nil
	el.Count = 0
}
