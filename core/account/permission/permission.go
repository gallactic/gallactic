package permission

import "github.com/gallactic/gallactic/core/account"

//------------------------------------------------------------------------------------------------
// Base permission references are like unix (the index is already bit shifted)
const (
	Root             account.Permissions = 1 << iota // 0x0001
	Send                                             // 0x0002
	Call                                             // 0x0004
	CreateContract                                   // 0x0008
	CreateAccount                                    // 0x0010
	Bond                                             // 0x0020
	ModifyPermission                                 // 0x0040
	CreateChain                                      // 0x0080
	InterChainTx

	Reserved
	None account.Permissions = 0
)

var (
	ZeroPermissions    account.Permissions
	AllPermissions     account.Permissions = (Reserved - 1)
	DefaultPermissions account.Permissions = Call | Send | CreateAccount | CreateContract
)

func EnsureValid(perm account.Permissions) bool {
	if (perm & ^AllPermissions) != 0 {
		return false
	}
	return true
}
