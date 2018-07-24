package burrow

import (
	"bytes"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	acm "github.com/hyperledger/burrow/acm"
	burrowBinary "github.com/hyperledger/burrow/binary"
	burrowCrypto "github.com/hyperledger/burrow/crypto"
	permission "github.com/hyperledger/burrow/permission"
)

type bState struct {
	st *state.State

	cache map[burrowCrypto.Address]*acm.MutableAccount
}

func (s bState) GetAccount(bAddr burrowCrypto.Address) (acm.Account, error) {
	if bAcc, ok := s.cache[bAddr]; ok {
		return bAcc, nil
	}

	addr := fromBurrowAddress(bAddr, false)
	acc := s.st.GetAccount(addr)

	if acc == nil {
		addr := fromBurrowAddress(bAddr, true)
		acc = s.st.GetAccount(addr)
	}

	if acc == nil {
		return acm.ConcreteAccount{
			Address: bAddr,
		}.Account(), nil
	}

	return toBurrowAccount(acc), nil
}

func (s bState) IterateAccounts(consumer func(acm.Account) (stop bool)) (stopped bool, err error) {
	return
}

func (s bState) UpdateAccount(updatedAccount acm.Account) error {
	return nil
}

func (s bState) RemoveAccount(bAddr burrowCrypto.Address) error {
	return nil
}

func (s bState) GetStorage(bAddr burrowCrypto.Address, key burrowBinary.Word256) (burrowBinary.Word256, error) {
	r := burrowBinary.Word256{}
	return r, nil
}
func (s bState) SetStorage(bAddr burrowCrypto.Address, key, value burrowBinary.Word256) error {
	return nil
}

func (s bState) IterateStorage(bAddr burrowCrypto.Address, consumer func(key, value binary.Word256) (stop bool)) (stopped bool, err error) {
	return false, nil
}

func toBurrowAccount(acc *account.Account) *acm.MutableAccount {

	bAddr := toBurrowAddress(acc.Address())
	bPerm := permission.AccountPermissions{
		Base: permission.BasePermissions{
			Perms:  permission.PermFlag(acc.Permissions()),
			SetBit: permission.PermFlag(acc.Permissions()),
		},
	}

	bs := [32]byte{}
	copy(bs[:], acc.Address().RawBytes())
	bPb, err := burrowCrypto.PublicKeyFromBytes(bs[:], burrowCrypto.CurveTypeEd25519)
	if err != nil {
		panic("cannot convert to burrow address")
	}

	bacc := &acm.ConcreteAccount{
		PublicKey:   bPb,
		Address:     bAddr,
		Balance:     acc.Balance(),
		Code:        acc.Code(),
		Sequence:    acc.Sequence(),
		Permissions: bPerm,
	}

	return bacc.MutableAccount()
}

func fromBurrowAccount(bAcc acm.MutableAccount) *account.Account {
	contract := len(bAcc.PublicKey().RawBytes()) == 0
	addr := fromBurrowAddress(bAcc.Address(), contract)
	perm := account.Permissions(bAcc.Permissions().Base.Perms)

	acc, _ := account.NewAccount(addr)
	acc.SetBalance(bAcc.Balance())
	acc.SetCode(bAcc.Code())
	acc.SetSequence(bAcc.Sequence())
	acc.SetPermissions(perm)
	//

	return acc

}

func toBurrowAddress(addr crypto.Address) burrowCrypto.Address {
	bAddr, err := burrowCrypto.AddressFromBytes(addr.RawBytes()[2:22])
	if err != nil {
		panic("cannot convert to burrow address")
	}
	return bAddr
}

func fromBurrowAddress(bAddr burrowCrypto.Address, contract bool) crypto.Address {

	var addr crypto.Address
	var err error
	if contract {
		addr, err = crypto.ContractAddress(bAddr.Bytes())
		if err != nil {
			panic("cannot convert to burrow address")
		}
	} else {
		if bytes.Equal(bAddr.Bytes(), crypto.GlobalAddress.RawBytes()[2:22]) {
			addr = crypto.GlobalAddress
		} else {
			addr, err = crypto.AccountAddress(bAddr.Bytes())
			if err != nil {
				panic("cannot convert to burrow address")
			}
		}
	}
	return addr
}
