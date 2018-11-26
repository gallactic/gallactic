package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyingSignature(t *testing.T) {
	msg := []byte("message")

	pb, pv := GenerateKey(nil)
	signer := NewAccountSigner(pv)
	sig, err := signer.Sign(msg)
	require.NoError(t, err)
	require.True(t, pb.Verify(msg, sig))
}
