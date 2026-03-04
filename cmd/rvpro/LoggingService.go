package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	jack "gopkg.in/natefinch/lumberjack.v2"
	"rvpro3/radarvision.com/internal/general"

	"rvpro3/radarvision.com/utils"
)

const logLevel = "log.level"
const logToConsole = "log.to.console"
const logFileDir = "log.file.dir"
const logFileName = "log.lile.name"
const logFileMaxSizeMB = "log.file.maxsizemb"
const logFileMaxAgeDays = "log.file.maxagedays"
const logFileMaxBackups = "log.file.maxbackups"

type LoggingService struct {
	FileDir      string
	FileName     string
	Level        string
	MaxSizeMB    int
	MaxAgeDays   int
	MaxBackups   int
	LogToConsole bool
}

func (l *LoggingService) InitFromSettings(settings *utils.Settings) {
	l.FileDir = settings.Basic.Get(logFileDir, "/media/SDLOGS/logs/system")
	l.FileName = settings.Basic.Get(logFileName, "rvm.log")
	l.Level = settings.Basic.Get(logLevel, "info")
	l.MaxSizeMB = settings.Basic.GetInt(logFileMaxSizeMB, 10)
	l.MaxAgeDays = settings.Basic.GetInt(logFileMaxAgeDays, 30)
	l.MaxBackups = settings.Basic.GetInt(logFileMaxBackups, 10)
	l.LogToConsole = settings.Basic.GetBool(logToConsole, false)
}

func (l *LoggingService) Start(state *utils.State, settings *utils.Settings) {
	if !general.ServiceHelper.ShouldStart(state, settings, l) {
		return
	}

	level := zerolog.InfoLevel

	var writers []io.Writer
	zerolog.TimeFieldFormat = utils.DisplayDateTimeMS

	if l.LogToConsole {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: utils.DisplayDateTimeMS,
		})
	}

	switch l.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	case "trace":
		level = zerolog.TraceLevel
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if l.FileDir != "" && l.FileName != "" {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        l.rollingAppender(settings),
			TimeFormat: utils.DisplayDateTimeMS,
		})
	}

	fmt.Println(level)
	mw := io.MultiWriter(writers...)
	logger := zerolog.New(mw).With().Timestamp().Logger()
	log.Logger = logger

	zerolog.SetGlobalLevel(level)
	log.WithLevel(level).Msg("Logging initialized")
}

func (l *LoggingService) rollingAppender(config *utils.Settings) io.Writer {
	if err := os.MkdirAll(l.FileDir, 0744); err != nil {
		log.Error().
			Err(err).
			Str("logFileDir", l.FileDir).
			Msg("Failed to create log directory")
		return nil
	}

	return &jack.Logger{
		Filename:   path.Join(l.FileDir, l.FileName),
		MaxSize:    l.MaxSizeMB,
		MaxAge:     l.MaxAgeDays,
		MaxBackups: l.MaxBackups,
	}
}

func (l *LoggingService) GetServiceName() string {
	return "Logging.Service"
}
