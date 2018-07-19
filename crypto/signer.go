package crypto

type Signer interface {
	Address() Address
	PublicKey() PublicKey
	Sign(msg []byte) (Signature, error)
}

type signer struct {
	pv PrivateKey
}

func NewAccountSigner(pv PrivateKey) Signer {
	return &signer{
		pv: pv,
	}
}

func (s *signer) Address() Address {
	return s.PublicKey().AccountAddress()
}

func (s *signer) PublicKey() PublicKey {
	return s.pv.PublicKey()
}

func (s *signer) Sign(msg []byte) (Signature, error) {
	return s.pv.Sign(msg)
}
