package permission

import "github.com/gallactic/gallactic/core/account"

//------------------------------------------------------------------------------------------------
// Base permission references are like unix (the index is already bit shifted)
const (
	// Send permits an account to issue a SendTx to transfer value from one account to another. Note that value can
	// still be transferred with a CallTx by specifying an Amount in the InputTx. Funding an account is the basic
	// prerequisite for an account to act in the system so is often used as a surrogate for 'account creation' when
	// sending to a unknown account - in order for this to be permitted the input account needs the CreateAccount
	// permission in addition.
	Send account.Permissions = 1 << iota // 0x0001
	// Call permits and account to issue a CallTx, which can be used to call (run) the code of an existing
	// account/contract (these are synonymous in Burrow/EVM). A CallTx can be used to create an account if it points to
	// a nil address - in order for an account to be permitted to do this the input (calling) account needs the
	// CreateContract permission in addition.
	Call // 0x0002
	// CreateContract permits the input account of a CallTx to create a new contract/account when CallTx.Address is nil
	// and permits an executing contract in the EVM to create a new contract programmatically.
	CreateContract // 0x0004
	// CreateAccount permits an input account of a SendTx to add value to non-existing (unfunded) accounts
	CreateAccount // 0x0008
	// Bond is a permission for making changes to the validator set
	Bond // 0x0010
	//
	ModifyPermission // 0x0020
	//
	CreateChain // 0x0040
	//
	InterChainTx // 0x080

	Reserved
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
