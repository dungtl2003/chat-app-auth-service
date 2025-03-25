package helper

import "log/slog"

// LoggerWrapper is a wrapper around slog.Logger, which is a structured logger. It is used to log messages in the application. If the logger is nil, the log messages will not be printed.
type LoggerWrapper struct {
	*slog.Logger
}

func NewLoggerWrapper(logger *slog.Logger) LoggerWrapper {
	return LoggerWrapper{logger}
}

func (l LoggerWrapper) Info(msg string, args ...any) {
	if l.Logger == nil {
		return
	}
	l.Logger.Info(msg, args...)
}

func (l LoggerWrapper) Error(msg string, args ...any) {
	if l.Logger == nil {
		return
	}
	l.Logger.Error(msg, args...)
}

func (l LoggerWrapper) Debug(msg string, args ...any) {
	if l.Logger == nil {
		return
	}
	l.Logger.Debug(msg, args...)
}
