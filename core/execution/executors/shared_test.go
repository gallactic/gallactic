package executors

/*
import (
	"testing"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
	"github.com/stretchr/testify/assert"
)

type fakeAccountGetter struct{}

func (fakeAccountGetter) GetAccount(addr crypto.Address) (*acm.Account, error) {
	if address == acm.GlobalAddress {
		globalAccount := acm.NewAccount(acm.GlobalAddress)
		globalAccount.SetPermissions(permission.Send | permission.Bond)
		return globalAccount, nil
	}

	return nil, nil
}

func TestHasPermission(t *testing.T) {
	var fakeGetter fakeAccountGetter

	acc := acm.NewAccountFromSecret("test")
	acc.SetPermissions(permission.Call)
	// Ensure we are falling through to global permissions on those bits not set
	assert.True(t, HasPermissions(fakeGetter, acc, permission.Call))
	assert.True(t, HasPermissions(fakeGetter, acc, permission.Send))
	assert.False(t, HasPermissions(fakeGetter, acc, permission.CreateAccount))

	acc.UnsetPermissions(permission.Call)
	assert.False(t, HasPermissions(fakeGetter, acc, permission.Call))

}
*/
