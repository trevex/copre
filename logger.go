package config

import (
	"fmt"
)

type LogLevel string

var (
	LogDebug LogLevel = "debug"
	LogInfo  LogLevel = "info"
	LogWarn  LogLevel = "warn"
	LogError LogLevel = "error"
)

type LogFunc func(logLevel LogLevel, msg string)

func StdoutLogger(minLogLevel LogLevel) LogFunc {
	allowedLogLevels := AllowedLogLevels(minLogLevel)
	return func(logLevel LogLevel, msg string) {
		if logLevelsContain(allowedLogLevels, logLevel) {
			fmt.Printf("[%s] %s\n", logLevel, msg)
		}
	}
}

func AllowedLogLevels(minLogLevel LogLevel) []LogLevel {
	allowedLogLevels := []LogLevel{LogDebug, LogInfo, LogWarn, LogError}
	minIndex := 0
	for i, logLevel := range allowedLogLevels {
		if logLevel == minLogLevel {
			minIndex = i
			break
		}
	}
	return allowedLogLevels[minIndex:]
}

func logLevelsContain(logLevels []LogLevel, expectedLogLevel LogLevel) bool {
	for _, logLevel := range logLevels {
		if logLevel == expectedLogLevel {
			return true
		}
	}
	return false
}
