package logger

import (
	"go.uber.org/zap"
)

var defaultLogger = NewLogger()

func Info(args ...any) {
	defaultLogger.sug.Info(args...)
}

func Infof(fmt string, args ...any) {
	defaultLogger.sug.Infof(fmt, args...)
}

func Debug(args ...any) {
	defaultLogger.sug.Debug(args...)
}

func Debugf(fmt string, args ...any) {
	defaultLogger.sug.Debugf(fmt, args...)
}

func Error(args ...any) {
	defaultLogger.sug.Error(args...)
}

func Errorf(fmt string, args ...any) {
	defaultLogger.sug.Errorf(fmt, args...)
}

func Warn(args ...any) {
	defaultLogger.sug.Warn(args...)
}

func Warnf(fmt string, args ...any) {
	defaultLogger.sug.Warnf(fmt, args...)
}

func Fatal(args ...any) {
	defaultLogger.sug.Fatal(args...)
}

func Fatalf(fmt string, args ...any) {
	defaultLogger.sug.Fatalf(fmt, args...)
}

func Panic(args ...any) {
	defaultLogger.sug.Panic(args...)
}

func Panicf(fmt string, args ...any) {
	defaultLogger.sug.Panicf(fmt, args...)
}

type Logger struct {
	sug *zap.SugaredLogger
}

func NewLogger() *Logger {
	log, _ := zap.NewDevelopment(zap.AddCallerSkip(1))
	sug := log.Sugar()

	return &Logger{
		sug: sug,
	}
}
