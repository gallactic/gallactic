package permission

import (
	"testing"

	"github.com/gallactic/gallactic/core/account"
	"github.com/stretchr/testify/assert"
)

func TestValidity(t *testing.T) {
	p1 := Reserved
	p2 := account.Permissions(0xFFFFFFFFFFFFFFFF)
	p3 := Call

	assert.False(t, EnsureValid(p1))
	assert.False(t, EnsureValid(p2))
	assert.True(t, EnsureValid(p3))
}

func TestModifying(t *testing.T) {
	p1 := Send
	p2 := Send
	p3 := Send | Call
	p2.Set(Call)
	p3.Unset(Call)

	assert.NotEqual(t, p1, p2)
	assert.Equal(t, p1, p3)
	assert.Equal(t, p1.IsSet(Call), false)
	assert.Equal(t, p2.IsSet(Call), true)
}

func TestAccountPermissions(t *testing.T) {
	acc := account.NewAccountFromSecret("PERM")
	acc.SetPermissions(Call)
	assert.Equal(t, acc.Permissions(), Call)
	acc.SetPermissions(CreateChain)
	assert.Equal(t, acc.Permissions(), Call|CreateChain)
	assert.Equal(t, acc.HasPermissions(InterChainTx), false)
	acc.UnsetPermissions(CreateChain)
	assert.Equal(t, acc.Permissions(), Call)
}
