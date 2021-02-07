package whatnot

import (
	"net/http"
)

/*
Health Check functions to use for Kubernetes Health-Check functions
*/

// Total Namespace Memory consumption

// Count of distinct Path Elements

// Percentage of currently locked Nodes

// Global HealthCheck poller
func Healthy() bool {
	return true
}

// attach this to your appropriate HTTP Healthcheck Endpoint
func HealthHandler(r *http.Request, w http.ResponseWriter) {

}

// is the number of backlogged lease attempts becoming unmanageable?
func waitLengthHealthCheck() bool {
	return true
}
