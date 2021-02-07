package whatnot

/*
Rate Limiting for Lease Acquisitions to prevent massive wait-state backlogs
*/

var WithRateLimit = func() optionName {
	return optionRateLimit
}

// is this lease request going to work?
// will it return in time?
func leaseLimitCheck() bool {
	return true
}
