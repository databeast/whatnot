// +build metrics
package whatnot

/*
Prometheus Metrics

build your projects with the build tag 'metrics' enabled to enable these

*/

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Current Locks

var lockCount prometheus.Counter
var currentLockCount prometheus.Gauge

func init() {
	lockCount = promauto.NewCounter(prometheus.CounterOpts{Name: "resuri_locks_since_start", Help: "Tptal number of distributed URI locks since boot"})
	currentLockCount = promauto.NewGauge(prometheus.GaugeOpts{Name: "resuri_current_lock_count", Help: "Current number of distributed URI locks"})
}

func incrementlocksSinceStart() {
	lockCount.Inc()
}

func incCurrentLock() {
	currentLockCount.Inc()
}

func decCurrentLock() {
	currentLockCount.Dec()
}
