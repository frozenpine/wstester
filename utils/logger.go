package utils

import "io"

// LogLevel log level
type LogLevel uint8

const (
	// TRACE trace level
	TRACE LogLevel = 1 << iota
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

// LevelAll all log levels
const LevelAll = TRACE | DEBUG | INFO | NOTICE | WARNING | ERROR | CRITICAL

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

	SetStdoutLevel(level LogLevel)
	DisableStdout()
	SetStderrLevel(level LogLevel)
	DisableStderr()
	SetSinker(sink io.Writer)
}

var root = logger{}

// Logln log line
func Logln(level LogLevel, args ...interface{}) {

}

// Logf log in format
func Logf(level LogLevel, format string, args ...interface{}) {

}

// Traceln trace log line
func Traceln(args ...interface{}) {
	Logln(TRACE, args...)
}

// Tracef trace log in format
func Tracef(format string, args ...interface{}) {
	Logf(TRACE, format, args...)
}

// Debugln debug log line
func Debugln(args ...interface{}) {
	Logln(DEBUG, args...)
}

// Debugf debug log in format
func Debugf(format string, args ...interface{}) {
	Logf(DEBUG, format, args...)
}

// Infoln info log line
func Infoln(args ...interface{}) {
	Logln(INFO, args...)
}

// Infof info log in format
func Infof(format string, args ...interface{}) {
	Logf(INFO, format, args...)
}

// Noticeln notice log line
func Noticeln(args ...interface{}) {
	Logln(NOTICE, args...)
}

// Noticef notice log in format
func Noticef(format string, args ...interface{}) {
	Logf(NOTICE, format, args...)
}

// Warningln warning log in line
func Warningln(args ...interface{}) {
	Logln(WARNING, args...)
}

// Warningf warning log in format
func Warningf(format string, args ...interface{}) {
	Logf(WARNING, format, args...)
}

// Errorln error log line
func Errorln(args ...interface{}) {
	Logln(ERROR, args...)
}

// Errorf error log in format
func Errorf(format string, args ...interface{}) {
	Logf(ERROR, format, args...)
}

// Criticalln error log line
func Criticalln(args ...interface{}) {
	Logln(CRITICAL, args...)
}

// Criticalf error log in format
func Criticalf(format string, args ...interface{}) {
	Logf(CRITICAL, format, args...)
}

// SetStdoutLevel set MAX log level for stdout
func SetStdoutLevel(level LogLevel) {

}

// DisableStdout disable output in stdout
func DisableStdout() {

}

// SetStderrLevel set MIN log level for stderr
func SetStderrLevel(level LogLevel) {

}

// DisableStderr disable output in stderr
func DisableStderr() {

}

// SetSinker set log sinker except stdout & stderr
func SetSinker(sink io.Writer) {

}

type logger struct {
	name     string
	parent   Logger
	chidls   []Logger
	sinker   io.Writer
	minLevel LogLevel
	maxLevel LogLevel
}

func (l *logger) inRange(lvl LogLevel) bool {
	return lvl >= l.minLevel && lvl <= l.maxLevel
}
