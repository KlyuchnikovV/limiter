package log

import "log/slog"

type Logger interface {
	With(...any) Logger

	Debug(string, ...any)
	Info(string, ...any)
	Error(string, ...any)
}

type SLogWrapper struct {
	*slog.Logger
}

func (log *SLogWrapper) With(args ...any) Logger {
	return WrapSLog(log.Logger.With(args...))
}

func WrapSLog(log *slog.Logger) Logger {
	return &SLogWrapper{
		Logger: log,
	}
}
