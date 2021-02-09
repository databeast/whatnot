package mutex

var mutexLogging MutexLog

// MutexLog allows you to implement/attach your own logger
type MutexLog interface {
	trace(msg string)
}

func (m *SmartMutex) trace(msg string) {
	//mutexLogging.trace(msg)
}

func dumpDeadlock() {

}
