package mutex

import (
	"io"
	"os"
	"sync"
	"time"
)

// Opts control how deadlock detection behaves.
// Options are supposed to be set once at a startup (say, when parsing flags).
var Opts = struct {
	// Mutex/rwmutex would work exactly as their sync counterparts
	// -- almost no runtime penalty, no deadlock detection if Disable == true.
	Disable bool

	// Would disable deadlockcheck order based deadlock detection if DisableDeadlockDetection == true.
	DisableDeadlockDetection bool

	// Waiting for a deadlockcheck for longer than DeadlockTimeout is considered a deadlock.
	// Ignored is DeadlockTimeout <= 0.
	DeadlockTimeout time.Duration

	// OnPotentialDeadlock is called each time a potential deadlock is detected -- either based on
	// deadlockcheck order or on deadlockcheck wait time.
	OnPotentialDeadlock func()

	// Will keep MaxMapSize deadlockcheck pairs (happens before // happens after) in the map.
	// The map resets once the threshold is reached.
	MaxMapSize int

	// Will dump stacktraces of all goroutines when inconsistent locking is detected.
	PrintAllCurrentGoroutines bool

	// Log out Lock Order Tracing
	Tracing bool

	mu *sync.Mutex // Protects the LogBuf.

	// Will print deadlock info to log buffer.
	LogBuf io.Writer
}{
	DeadlockTimeout: time.Second * 30,
	OnPotentialDeadlock: func() {
		os.Exit(2)
	},
	MaxMapSize: 1024 * 64,
	mu:         &sync.Mutex{},
	LogBuf:     os.Stderr,
}
