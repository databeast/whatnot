package access

// Identity is where your application tracks the Security Identity (a user, etc)
// that is accessing/leasing/locking a namespace element
type Identity struct {
	roles []*Role
}

type id string

type identities map[id]*Identity

func RegisterNewIdentity() {

}
