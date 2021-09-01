package whatnot

import (
	"math/rand"
	"time"
)

// random generator for internal IDs
var randid = rand.New(rand.NewSource(time.Now().UnixNano()))
