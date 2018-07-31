package permission

import "github.com/gallactic/gallactic/core/account"

//------------------------------------------------------------------------------------------------
// Base permission references are like unix (the index is already bit shifted)
const (
	None             account.Permissions = 0
	Send             account.Permissions = 1 << iota // 0x0001
	Call                                             // 0x0002
	Bond                                             // 0x0004
	CreateContract                                   // 0x0008
	CreateAccount                                    // 0x0010
	ModifyPermission                                 // 0x0020
	CreateChain                                      // 0x0040
	InterChainTx                                     // 0x0080

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
