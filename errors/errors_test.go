package e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCode(t *testing.T) {
	err1 := Error(ErrInsufficientFunds)
	require.Equal(t, Code(err1), ErrInsufficientFunds)

	err2 := fmt.Errorf("Nope")
	require.Equal(t, Code(err2), ErrGeneric)
}
