package main

import (
	"github.com/sirupsen/logrus"
	"os"
)

var (
	Logger = newLogger(logrus.DebugLevel)
)

type ILogger interface {
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})

	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Fatal(v ...interface{})
}

type defaultLogger struct {
}

func newLogger(debugLevel logrus.Level) ILogger {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(debugLevel)
	return &defaultLogger{}
}

func (*defaultLogger) Debugf(format string, v ...interface{}) {
	logrus.Debugf(format, v...)
}

func (*defaultLogger) Infof(format string, v ...interface{}) {
	logrus.Infof(format, v...)
}

func (*defaultLogger) Warnf(format string, v ...interface{}) {
	logrus.Warnf(format, v...)
}
func (*defaultLogger) Errorf(format string, v ...interface{}) {
	logrus.Errorf(format, v...)
}

func (*defaultLogger) Fatalf(format string, v ...interface{}) {
	logrus.Fatalf(format, v...)
}

func (*defaultLogger) Debug(v ...interface{}) {
	logrus.Debug(v...)
}

func (*defaultLogger) Info(v ...interface{}) {
	logrus.Info(v...)
}

func (*defaultLogger) Warn(v ...interface{}) {
	logrus.Warn(v...)
}
func (*defaultLogger) Error(v ...interface{}) {
	logrus.Error(v...)
}

func (*defaultLogger) Fatal(v ...interface{}) {
	logrus.Fatal(v...)
}
