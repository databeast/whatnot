package whatnot

var namespaceLogging Logger

// Logger allows you to implement/attach your own logger
type Logger interface {
	Debug(msg string)
	Info(msg string)
}

func (m Namespace) log() {

}

type nilLogger struct{}

func (n nilLogger) Debug(msg string) {}

func (n nilLogger) Info(msg string) {}
