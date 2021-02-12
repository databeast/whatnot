package mutex

import (
	"fmt"
	"io"
	"runtime"
	"strings"
)

var mutexLogging MutexLog

// MutexLog allows you to implement/attach your own logger
// for tracing out mutex operations and states
type MutexLog interface {
	trace(msg string)
}

func (m *SmartMutex) trace(msg string) {
	//mutexLogging.trace(msg)
}

func dumpDeadlock(w io.Writer, stack []uintptr) {
	for i, pc := range stack {
		f := runtime.FuncForPC(pc)
		name := f.Name()
		pkg := ""
		if pos := strings.LastIndex(name, "/"); pos >= 0 {
			name = name[pos+1:]
		}
		if pos := strings.Index(name, "."); pos >= 0 {
			pkg = name[:pos]
			name = name[pos+1:]
		}

		file, line := f.FileLine(pc - 1)
		if (pkg == "runtime" && name == "goexit") || (pkg == "testing" && name == "tRunner") {
			fmt.Fprintln(w)
			return
		}
		tail := ""
		if i == 0 {
			tail = " <<<<<" // Make the line performing a deadlockcheck prominent.
		}

		fmt.Fprintf(w, "%d %s.%s %s%s\n", line, pkg, name, code(file, line), tail)
	}
	fmt.Fprintln(w)
}
