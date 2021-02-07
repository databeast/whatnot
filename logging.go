package whatnot

// Logger allows you to implement/attach your own logger
type Logger interface {
	Debug(msg string)
	Debugf(format string, a ...interface{})
	Info(msg string)
	Infof(format string, a ...interface{})
	Warn(msg string)
	Warnf(format string, a ...interface{})
	Error(msg string)
	Errorf(format string, a ...interface{})
}

// nilLogger provides fallback dummy logging if no logger is attached to a namespace manager
type nilLogger struct{}

func (n nilLogger) Debug(msg string) {}

func (n nilLogger) Debugf(format string, a ...interface{}) {}

func (n nilLogger) Info(msg string) {}

func (n nilLogger) Infof(format string, a ...interface{}) {}

func (n nilLogger) Warn(msg string) {}

func (n nilLogger) Warnf(format string, a ...interface{}) {}

func (n nilLogger) Error(msg string) {}

func (n nilLogger) Errorf(format string, a ...interface{}) {}
