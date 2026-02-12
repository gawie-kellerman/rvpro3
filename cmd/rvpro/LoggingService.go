package main

import (
	"io"
	"os"
	"path"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	jack "gopkg.in/natefinch/lumberjack.v2"
	"rvpro3/radarvision.com/utils"
)

const logLevel = "Log.Level"
const logToConsole = "Log.To.Console"
const logFileDir = "Log.File.Dir"
const logFileName = "Log.File.Name"
const logFileMaxSizeMB = "Log.File.MaxSizeMB"
const logFileMaxAgeDays = "Log.File.MaxAgeDays"
const logFileMaxBackups = "Log.File.MaxBackups"

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

	if fileDir != "" && fileName != "" {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        l.rollingAppender(config),
			TimeFormat: utils.DisplayDateTimeMS,
		})
	}

	mw := io.MultiWriter(writers...)
	logger := zerolog.New(mw).With().Timestamp().Logger()

	log.Logger = logger

	log.Info().Msg("Logging initialized")
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
