package whatnot

type NamespaceManagerOpt interface {
	opt()
}

type optionName string

const (
	optionDiscoverGossip optionName = "gossip cluster discovery"
	optionSyncRaft       optionName = "raft cluster synchronization"
	optionTrace          optionName = "trace mutex locks"
	optionBreak          optionName = "break mutex deadlocking"
	optionAcls           optionName = "enable element permissions"
	optionRateLimit      optionName = "lease rate limiting"
	optionLogger         optionName = "custom log output"
)

type ManagerOption interface {
	apply(manager *NameSpaceManager) (err error)
	name() optionName
}

type managerOptionFunc func() optionName

var WithGossip managerOptionFunc = func() optionName {
	return optionDiscoverGossip
}

var WithRaft managerOptionFunc = func() optionName {
	return optionSyncRaft
}

var WithTrace managerOptionFunc = func() optionName {
	return optionTrace
}

var WithDeadlockBreak managerOptionFunc = func() optionName {
	return optionBreak
}

var WithAcls managerOptionFunc = func() optionName {
	return optionAcls
}

func (f managerOptionFunc) apply(manager *NameSpaceManager) (err error) {
	return
}

func (f managerOptionFunc) name() optionName {
	return f()
}

// WithLogger attaches a Logger to the Namespace manager
// allowing you to insert your own logging solution into Whatnot
type WithLogger struct {
	l Logger
}

func (w WithLogger) name() optionName {
	return optionLogger
}

func (w WithLogger) apply(manager *NameSpaceManager) (err error) {
	if w.l == nil {
		return newConfigError("no logger passed in withlogger config option")
	}
	w.l.Warn("replacing whatnot logging target")
	whatlogger = w.l
	return nil
}
