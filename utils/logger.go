package utils

import "io"

// LogLevel log level
type LogLevel uint8

const (
	// TRACE trace level
	TRACE LogLevel = iota
	// DEBUG debug level
	DEBUG
	// INFO info level
	INFO
	// NOTICE notice level
	NOTICE
	// WARNING warning level
	WARNING
	// ERROR error level
	ERROR
	// CRITICAL critical level
	CRITICAL
)

// Logger logger interface
type Logger interface {
	Logln(level LogLevel, args ...interface{})
	Logf(level LogLevel, format string, args ...interface{})
	Traceln(args ...interface{})
	Tracef(format string, args ...interface{})
	Debugln(args ...interface{})
	Debugf(format string, args ...interface{})
	Infoln(args ...interface{})
	Infof(format string, args ...interface{})
	Noticeln(args ...interface{})
	Noticef(format string, args ...interface{})
	Warningln(args ...interface{})
	Warningf(format string, args ...interface{})
	Errorln(args ...interface{})
	Errorf(format string, args ...interface{})
	Criticalln(args ...interface{})
	Criticalf(format string, args ...interface{})
	Panicln(args ...interface{})
	Panicf(format string, args ...interface{})

	SetStdoutLevel(level LogLevel)
	DisableStdout()
	SetStderrLevel(level LogLevel)
	DisableStderr()
	SetSinker(sink io.Writer)
}
