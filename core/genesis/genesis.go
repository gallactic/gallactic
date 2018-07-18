package genesis

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"time"

	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/validator"
	"github.com/gallactic/gallactic/crypto"
)

// How many bytes to take from the front of the Genesis hash to append
// to the ChainName to form the ChainID. The idea is to avoid some classes
// of replay attack between chains with the same name.
const shortHashSuffixBytes = 3

//------------------------------------------------------------
// core types for a genesis definition

type genAccount struct {
	Address     crypto.Address      `json:"address"`
	Balance     uint64              `json:"balance"`
	Permissions account.Permissions `json:"permissions"`
}

type genValidator struct {
	Address   crypto.Address   `json:"address"`
	Stake     uint64           `json:"stake"`
	PublicKey crypto.PublicKey `json:"publicKey"`
}

//------------------------------------------------------------
// Genesis is stored in the state database
type Genesis struct {
	data genesisData
}

type genesisData struct {
	ChainName         string              `json:"chainName"`
	GenesisTime       time.Time           `json:"genesisTime"`
	GlobalPermissions account.Permissions `json:"globalPermission"`
	MaximumPower      int                 `json:"maximumPower"`
	Accounts          []genAccount        `json:"accounts"`
	Validators        []genValidator      `json:"validators"`
}

func (gen Genesis) ChainName() string {
	return gen.data.ChainName
}

func (gen Genesis) GenesisTime() time.Time {
	return gen.data.GenesisTime
}

func (gen Genesis) GlobalPermissions() account.Permissions {
	return gen.data.GlobalPermissions
}

func (gen Genesis) Hash() []byte {
	genesisDocBytes, err := gen.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("could not create hash of Genesis: %v", err))
	}
	hasher := sha256.New()
	hasher.Write(genesisDocBytes)
	return hasher.Sum(nil)
}

func (gen Genesis) ShortHash() []byte {
	return gen.Hash()[:shortHashSuffixBytes]
}

func (gen Genesis) ChainID() string {
	return fmt.Sprintf("%s-%X", gen.data.ChainName, gen.ShortHash())
}

func (gen Genesis) Accounts() []*account.Account {
	accounts := make([]*account.Account, 0, len(gen.data.Accounts))
	for _, genAccount := range gen.data.Accounts {
		account, err := account.NewAccount(genAccount.Address)
		if err == nil {
			account.SetPermissions(genAccount.Permissions)
			account.AddToBalance(genAccount.Balance)

			accounts = append(accounts, account)
		}
	}

	return accounts
}

func (gen Genesis) Validators() []*validator.Validator {
	validators := make([]*validator.Validator, 0, len(gen.data.Validators))
	for _, genValidator := range gen.data.Validators {
		validator := validator.NewValidator(genValidator.PublicKey, genValidator.Stake, 0)

		validators = append(validators, validator)
	}

	return validators
}

func (gen Genesis) MaximumPower() int {
	if gen.data.MaximumPower < len(gen.data.Validators) {
		return len(gen.data.Validators)
	}

	return gen.data.MaximumPower
}

//------------------------------------------------------------
// Make genesis state from file

func (gen Genesis) MarshalJSON() ([]byte, error) {
	return json.Marshal(&gen.data)
}

func (gen *Genesis) UnmarshalJSON(bs []byte) error {
	err := json.Unmarshal(bs, &gen.data)
	if err != nil {
		return err
	}
	return nil
}

func makeGenesisAccount(account *account.Account) genAccount {
	return genAccount{
		Address:     account.Address(),
		Balance:     account.Balance(),
		Permissions: account.Permissions(),
	}
}

func makeGenesisValidator(validator *validator.Validator) genValidator {
	return genValidator{
		PublicKey: validator.PublicKey(),
		Address:   validator.Address(),
		Stake:     validator.Balance(),
	}
}

// MakeGenesisDoc takes a chainName and a slice of pointers to Account,
// and a slice of pointers to Validator to construct a Genesis, or returns an error on
// failure.
func MakeGenesisDoc(chainName string, genesisTime time.Time, globalPerms account.Permissions,
	accounts []*account.Account, validators []*validator.Validator) *Genesis {

	// Establish deterministic order of accounts by address so we obtain identical Genesis
	// from identical input
	sort.SliceStable(accounts, func(i, j int) bool {
		return bytes.Compare(accounts[i].Address().RawBytes(), accounts[j].Address().RawBytes()) < 0
	})

	sort.SliceStable(validators, func(i, j int) bool {
		return bytes.Compare(validators[i].Address().RawBytes(), validators[j].Address().RawBytes()) < 0
	})

	// copy slice of pointers to accounts
	genAccounts := make([]genAccount, 0, len(accounts))
	for _, account := range accounts {
		genAccount := makeGenesisAccount(account)

		genAccounts = append(genAccounts, genAccount)
	}

	// copy slice of pointers to validators
	genValidators := make([]genValidator, 0, len(validators))
	for _, validator := range validators {
		genValidator := makeGenesisValidator(validator)

		genValidators = append(genValidators, genValidator)
	}

	return &Genesis{
		data: genesisData{
			ChainName:         chainName,
			GenesisTime:       genesisTime,
			GlobalPermissions: globalPerms,
			Accounts:          genAccounts,
			Validators:        genValidators,
		},
	}
}

func LoadFromFile(file string) (*Genesis, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var gen Genesis
	if err := json.Unmarshal(dat, &gen); err != nil {
		return nil, err
	}
	return &gen, nil
}
