package whatnot

import (
	"testing"
)

func TestLeaseCreationAndExpiration(t *testing.T) {
	t.Run("Create a new lease", createNewLeaseOnPathElement)

}

func createNewLeaseOnPathElement(t *testing.T) {
	t.Log("Creating a new lease on an existing Path Element")

}

func leaseExpiresAsExpected(t *testing.T) {
	t.Log("Checking that lease expires at expected duration")

}

func leaseAcceptsCancelation(t *testing.T) {
	t.Log("Creating a lease and then cancelling it prematurely")

}
