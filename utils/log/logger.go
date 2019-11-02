package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	errorLogger *zap.SugaredLogger
	atom        zap.AtomicLevel
)

// TimeEncoder time encoder
func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func init() {
	atom = zap.NewAtomicLevel()
	fileName := fmt.Sprintf("./log/%04d-%02d-%02d.log", time.Now().Year(), time.Now().Month(), time.Now().Day())
	syncWriter := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    1024, // megabytes
		MaxBackups: 10,
		MaxAge:     30, // days
		LocalTime:  true,
		Compress:   true,
	}))
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		syncWriter,
		atom,
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	errorLogger = logger.Sugar()
}

// SetLogLevel set log level
func SetLogLevel(level zapcore.Level) {
	atom.SetLevel(level)
}

// Debug debug log line
func Debug(args ...interface{}) {
	errorLogger.Debug(args...)
}

// Debugf debug log in format
func Debugf(template string, args ...interface{}) {
	errorLogger.Debugf(template, args...)
}

// Info info log line
func Info(args ...interface{}) {
	errorLogger.Info(args...)
}

// Infof info log in format
func Infof(template string, args ...interface{}) {
	errorLogger.Infof(template, args...)
}

// Warn warning log line
func Warn(args ...interface{}) {
	errorLogger.Warn(args...)
}

// Warnf warning log in format
func Warnf(template string, args ...interface{}) {
	errorLogger.Warnf(template, args...)
}

// Error error log line
func Error(args ...interface{}) {
	errorLogger.Error(args...)
}

// Errorf error log in format
func Errorf(template string, args ...interface{}) {
	errorLogger.Errorf(template, args...)
}

// DPanic panic log line in debug mode
func DPanic(args ...interface{}) {
	errorLogger.DPanic(args...)
}

// DPanicf panic log in format in debug mode
func DPanicf(template string, args ...interface{}) {
	errorLogger.DPanicf(template, args...)
}

// Panic panic log line
func Panic(args ...interface{}) {
	errorLogger.Panic(args...)
}

// Panicf panic log in format
func Panicf(template string, args ...interface{}) {
	errorLogger.Panicf(template, args...)
}

// Fatal fatal log line
func Fatal(args ...interface{}) {
	errorLogger.Fatal(args...)
}

// Fatalf fatal log in format
func Fatalf(template string, args ...interface{}) {
	errorLogger.Fatalf(template, args...)
}
