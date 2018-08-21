package crypto

// Signatory contains signature and PublicKey to identify the signer
type Signatory struct {
	PublicKey PublicKey `json:"publicKey"`
	Signature Signature `json:"signature"`
}
