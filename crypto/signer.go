package crypto

type Signer interface {
	Address() Address
	PublicKey() PublicKey
	PrivateKey() PrivateKey
	Sign(msg []byte) (Signature, error)
}
