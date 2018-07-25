package proposal

import "github.com/gallactic/gallactic/crypto"

type signatory struct {
	PublicKey crypto.PublicKey
	Signature crypto.Signature
}

type Proposal struct {
	Genesis     *Genesis    `json:"genesis"`
	Signatories []signatory `json:"signatories,omitempty"`
}
