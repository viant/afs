package option

type Logger struct {
	logf func(format string, args ...interface{})
}

func (l *Logger) Logf(format string, args ...interface{}) {
	if l.logf == nil {
		return
	}
	l.logf(format, args...)
}

func WithLogger(logf func(format string, args ...interface{})) *Logger {
	return &Logger{logf: logf}
}
