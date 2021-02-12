package access

// Role is where your application tracks what capability a given identity is using
// to access/lease/lock a namespace resource
type Role struct {
	Name        string
	UsageHook   RoleUsageHookFunc
	Permissions *PermissionSet
}

// RoleUsageHookFunc declares a function provided by your code
// which will be called (as a separate goroutine) when
// a declared role is used to grant access to a Namespace
// Path Element
type RoleUsageHookFunc func()
