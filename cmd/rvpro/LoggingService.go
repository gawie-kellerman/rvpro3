package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	jack "gopkg.in/natefinch/lumberjack.v2"
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
}

func (l *LoggingService) SetupDefaults(config *utils.Settings) {
	config.SetSettingAsStr(logLevel, "info")
	config.SetSettingAsBool(logToConsole, false)
	config.SetSettingAsStr(logFileDir, "/media/SDLOGS/logs/system")
	config.SetSettingAsStr(logFileName, "rvm.log")
	config.SetSettingAsInt(logFileMaxSizeMB, 10)
	config.SetSettingAsInt(logFileMaxAgeDays, 30)
	config.SetSettingAsInt(logFileMaxBackups, 10)
}

func (l *LoggingService) SetupAndStart(state *utils.State, config *utils.Settings) {
	level := zerolog.InfoLevel

	var writers []io.Writer
	zerolog.TimeFieldFormat = utils.DisplayDateTimeMS

	if config.GetSettingAsBool(logToConsole) {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: utils.DisplayDateTimeMS,
		})
	}

	fileDir := config.GetSettingAsStr(logFileDir)
	fileName := config.GetSettingAsStr(logFileName)
	logLevelStr := config.GetSettingAsStr(logLevel)
	switch logLevelStr {
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

	if fileDir != "" && fileName != "" {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        l.rollingAppender(config),
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
	fileDir := config.GetSettingAsStr(logFileDir)
	fileName := config.GetSettingAsStr(logFileName)

	gc := &utils.GlobalSettings

	if err := os.MkdirAll(fileDir, 0744); err != nil {
		log.Error().
			Err(err).
			Str("logFileDir", fileDir).
			Msg("Failed to create log directory")
		return nil
	}

	return &jack.Logger{
		Filename:   path.Join(fileDir, fileName),
		MaxSize:    gc.GetSettingAsInt(logFileMaxSizeMB),
		MaxAge:     gc.GetSettingAsInt(logFileMaxAgeDays),
		MaxBackups: gc.GetSettingAsInt(logFileMaxBackups),
	}
}

func (l *LoggingService) GetServiceName() string {
	return "Logging.Service"
}

func (l *LoggingService) GetServiceNames() []string {
	return nil
}
