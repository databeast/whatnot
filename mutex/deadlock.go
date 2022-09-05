package mutex

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime"
	"sync"
	"time"

	"github.com/petermattis/goid"
)

const deadlockdumpheader = "POTENTIAL DEADLOCK:"

// THIS IS THE THING THAT DOES DEADLOCK ANALYSIS
func deadlockcheck(lockFn func(), ptr interface{}) {
	if Opts.Disable {
		lockFn()
		return
	}
	preLock(4, ptr)
	if Opts.DeadlockTimeout <= 0 {
		lockFn()
	} else {
		// Begin goroutine to monitor potential Deadlock state
		ch := make(chan struct{})
		go func() {
			for {
				t := time.NewTimer(Opts.DeadlockTimeout)
				defer t.Stop() // This runs after the closure finishes, but it's OK.
				select {
				case <-t.C:
					lo.mu.Lock()
					prev, ok := lo.cur[ptr]
					if !ok {
						lo.mu.Unlock()
						break // Nobody seems to be holding the deadlockcheck, try again.
					}
					Opts.mu.Lock()
					fmt.Fprintln(Opts.LogBuf, deadlockdumpheader)
					fmt.Fprintln(Opts.LogBuf, "Previous place where the lock was grabbed")
					fmt.Fprintf(Opts.LogBuf, "goroutine %v lock %p\n", prev.gid, ptr)
					dumpDeadlock(Opts.LogBuf, prev.stack)
					fmt.Fprintln(Opts.LogBuf, "Have been trying to lock it again for more than", Opts.DeadlockTimeout)
					fmt.Fprintf(Opts.LogBuf, "goroutine %v lock %p\n", goid.Get(), ptr)
					dumpDeadlock(Opts.LogBuf, callers(2))
					stacks := stacks()
					grs := bytes.Split(stacks, []byte("\n\n"))
					for _, g := range grs {
						if goid.ExtractGID(g) == prev.gid {
							fmt.Fprintln(Opts.LogBuf, "Here is what goroutine", prev.gid, "doing now")
							Opts.LogBuf.Write(g)
							fmt.Fprintln(Opts.LogBuf)
						}
					}
					lo.other(ptr)
					if Opts.PrintAllCurrentGoroutines {
						fmt.Fprintln(Opts.LogBuf, "All current goroutines:")
						Opts.LogBuf.Write(stacks)
					}
					fmt.Fprintln(Opts.LogBuf)
					if buf, ok := Opts.LogBuf.(*bufio.Writer); ok {
						buf.Flush()
					}
					Opts.mu.Unlock()
					lo.mu.Unlock()
					Opts.OnPotentialDeadlock()
					<-ch
					return
				case <-ch:
					return
				}
			}
		}()
		lockFn()
		postLock(4, ptr)
		close(ch)
		return
	}
	postLock(4, ptr)
}

type stackGID struct {
	stack []uintptr
	gid   int64
}

type beforeAfter struct {
	before interface{}
	after  interface{}
}

type ss struct {
	before []uintptr
	after  []uintptr
}

func (l *lockOrder) preLock(skip int, p interface{}) {
	if Opts.DisableDeadlockDetection {
		return
	}
	stack := callers(skip)
	gid := goid.Get()
	l.mu.Lock()
	for b, bs := range l.cur {
		if b == p {
			if bs.gid == gid {
				Opts.mu.Lock()
				fmt.Fprintln(Opts.LogBuf, deadlockdumpheader, "Recursive locking:")
				fmt.Fprintf(Opts.LogBuf, "current goroutine %d deadlockcheck %p\n", gid, b)
				dumpDeadlock(Opts.LogBuf, stack)
				fmt.Fprintln(Opts.LogBuf, "Previous place where the deadlockcheck was grabbed (same goroutine)")
				dumpDeadlock(Opts.LogBuf, bs.stack)
				l.other(p)
				if buf, ok := Opts.LogBuf.(*bufio.Writer); ok {
					buf.Flush()
				}
				Opts.mu.Unlock()
				Opts.OnPotentialDeadlock()
			}
			continue
		}
		if bs.gid != gid { // We want locks taken in the same goroutine only.
			continue
		}
		if s, ok := l.order[beforeAfter{p, b}]; ok {
			Opts.mu.Lock()
			fmt.Fprintln(Opts.LogBuf, deadlockdumpheader, "Inconsistent locking. saw this ordering in one goroutine:")
			fmt.Fprintln(Opts.LogBuf, "happened before")
			dumpDeadlock(Opts.LogBuf, s.before)
			fmt.Fprintln(Opts.LogBuf, "happened after")
			dumpDeadlock(Opts.LogBuf, s.after)
			fmt.Fprintln(Opts.LogBuf, "in another goroutine: happened before")
			dumpDeadlock(Opts.LogBuf, bs.stack)
			fmt.Fprintln(Opts.LogBuf, "happened after")
			dumpDeadlock(Opts.LogBuf, stack)
			l.other(p)
			fmt.Fprintln(Opts.LogBuf)
			if buf, ok := Opts.LogBuf.(*bufio.Writer); ok {
				buf.Flush()
			}
			Opts.mu.Unlock()
			Opts.OnPotentialDeadlock()
		}
		l.order[beforeAfter{b, p}] = ss{bs.stack, stack}
		if len(l.order) == Opts.MaxMapSize { // Reset the map to keep memory footprint bounded.
			l.order = map[beforeAfter]ss{}
		}
	}
	l.mu.Unlock()
}

func callers(skip int) []uintptr {
	s := make([]uintptr, 50) // Most relevant context seem to appear near the top of the stack.
	return s[:runtime.Callers(2+skip, s)]
}

var fileSources struct {
	sync.Mutex
	lines map[string][][]byte
}

// Reads souce file lines from disk if not cached already.
func getSourceLines(file string) [][]byte {
	fileSources.Lock()
	defer fileSources.Unlock()
	if fileSources.lines == nil {
		fileSources.lines = map[string][][]byte{}
	}
	if lines, ok := fileSources.lines[file]; ok {
		return lines
	}
	text, _ := ioutil.ReadFile(file)
	fileSources.lines[file] = bytes.Split(text, []byte{'\n'})
	return fileSources.lines[file]
}

func code(file string, line int) string {
	lines := getSourceLines(file)
	// lines are 1 based.
	if line >= len(lines) || line <= 0 {
		return "???"
	}
	return "{ " + string(bytes.TrimSpace(lines[line-1])) + " }"
}

// Stacktraces for all goroutines.
func stacks() []byte {
	buf := make([]byte, 1024*16)
	for {
		n := runtime.Stack(buf, true)
		if n < len(buf) {
			return buf[:n]
		}
		buf = make([]byte, 2*len(buf))
	}
}
