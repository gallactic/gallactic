package crypto

type Signer interface {
	Address() Address
	PublicKey() PublicKey
	Sign(msg []byte) (Signature, error)
	SignWithoutHash(msg []byte) (Signature, error) // Sign the message without hashing it
}

type signer struct {
	address    Address
	publicKey  PublicKey
	privateKey PrivateKey
}

func NewAccountSigner(pv PrivateKey) Signer {
	return &signer{
		privateKey: pv,
		publicKey:  pv.PublicKey(),
		address:    pv.PublicKey().AccountAddress(),
	}
}

func NewValidatorSigner(pv PrivateKey) Signer {
	return &signer{
		privateKey: pv,
		publicKey:  pv.PublicKey(),
		address:    pv.PublicKey().ValidatorAddress(),
	}
}

func (s *signer) Address() Address {
	return s.address
}

func (s *signer) PublicKey() PublicKey {
	return s.publicKey
}

func (s *signer) Sign(msg []byte) (Signature, error) {
	return s.privateKey.Sign(msg)
}

func (s *signer) SignWithoutHash(msg []byte) (Signature, error) {
	return s.privateKey.SignWithoutHash(msg)
}
