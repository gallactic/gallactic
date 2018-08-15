package proposal

import (
	"fmt"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
)

type Proposal struct {
	Genesis     *Genesis           `json:"genesis"`
	Signatories []crypto.Signatory `json:"signatories,omitempty"`
}

func (p *Proposal) Verify() error {
	/* check for signatories */
	if len(p.Signatories) == 0 {
		return e.Errorf(e.ErrInvalidSignature, "propsal contains no  signatories")
	}

	/* genesis accounts */
	accounts := p.Genesis.Accounts()

	/* genesis validator accounts */
	validators := p.Genesis.Validators()

	/* genesis hash */
	signHash := p.Genesis.Hash()

	// -1 for global account
	if len(p.Signatories) != len(accounts)+len(validators)-1 {
		return e.Errorf(e.ErrInvalidSignature, "Invalid signature")
	}

	/* check accounts signature */
	found := false
	for _, acc := range accounts {
		if acc.Address().EqualsTo(crypto.GlobalAddress) {
			continue
		}

		for _, sig := range p.Signatories {
			if acc.Address().EqualsTo(sig.PublicKey.AccountAddress()) {
				found = true
				if !sig.PublicKey.Verify(signHash, sig.Signature) {
					return fmt.Errorf("Cannot verify the signature for account %s", acc.Address())
				}
			}
		}
		if !found {
			return fmt.Errorf("No signature for account %s", acc.Address())
		}
	}

	/* check validator signature */
	for _, val := range validators {
		for _, sig := range p.Signatories {
			if val.Address().EqualsTo(sig.PublicKey.ValidatorAddress()) {
				found = true
				if !sig.PublicKey.Verify(signHash, sig.Signature) {
					return fmt.Errorf("Cannot verify the signature for valiator %s", val.Address())
				}
			}

		}
		if !found {
			return fmt.Errorf("No signature for validator %s", val.Address())
		}
	}

	return nil
}
