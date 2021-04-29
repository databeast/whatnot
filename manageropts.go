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
	optionPruning		 optionName = "unused element pruning"
)

type ManagerOption interface {
	apply(manager *NameSpaceManager) (err error)
	name() optionName
}

type managerOptionFunc func() optionName

// WithGossip enables Gossip protocol Cluster member discovery of other instances running
// the whatNot gRPC connector
var WithGossip managerOptionFunc = func() optionName {
	return optionDiscoverGossip
}

// WithRaft enables Raft Quorum synchromization - improving cluster accuracy at a slight speed and bandwith cost
var WithRaft managerOptionFunc = func() optionName {
	return optionSyncRaft
}

// WithTrace enables extended tracing of Resource Locking and Wait Queues
var WithTrace managerOptionFunc = func() optionName {
	return optionTrace
}

// WithDeadlockBreak turns on Whatnot's Self-healing breaking of Mutex Deadlocks
var WithDeadlockBreak managerOptionFunc = func() optionName {
	return optionBreak
}

// WithAcls turns on Whatnot's Access Control Management on individual Path Elements
var WithAcls managerOptionFunc = func() optionName {
	return optionAcls
}

// WithPruning turns on Whatnot's automatic pruning of Unused PathElement Tree sections after
// they remain unused for a given amount of time
var WithPruning managerOptionFunc = func() optionName {
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
