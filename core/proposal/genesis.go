package proposal

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
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

// core types for a genesis definition

type genAccount struct {
	Address     crypto.Address      `json:"address"`
	Balance     uint64              `json:"balance"`
	Permissions account.Permissions `json:"permissions,omitempty"`
}

type globalAccount struct {
	Balance     uint64              `json:"balance"`
	Permissions account.Permissions `json:"permissions"`
	Code        []byte              `json:"code,omitempty"`
}

type genContract struct {
	Address     crypto.Address      `json:"address"`
	Code        []byte              `json:"code"`
	Permissions account.Permissions `json:"permissions"`
}

type genValidator struct {
	Address   crypto.Address   `json:"address"`
	Stake     uint64           `json:"stake"`
	PublicKey crypto.PublicKey `json:"publicKey"`
}

// Genesis is stored in the state database
type Genesis struct {
	data genesisData
}

type genesisData struct {
	ChainName     string         `json:"chainName"`
	GenesisTime   time.Time      `json:"genesisTime"`
	MaximumPower  int            `json:"maximumPower"`
	SortitionFee  int            `json:"sortitionFee"`
	GlobalAccount globalAccount  `json:"global"`
	Accounts      []genAccount   `json:"accounts"`
	Contracts     []genContract  `json:"contracts"`
	Validators    []genValidator `json:"validators"`
}

func (gen *Genesis) Hash() []byte {
	genesisDocBytes, err := gen.MarshalJSON()
	if err != nil {
		panic(fmt.Errorf("could not create hash of Genesis: %v", err))
	}
	hasher := sha256.New()
	hasher.Write(genesisDocBytes)
	return hasher.Sum(nil)
}

func (gen *Genesis) ShortHash() []byte {
	return gen.Hash()[:shortHashSuffixBytes]
}

func (gen *Genesis) ChainName() string {
	return gen.data.ChainName
}

func (gen *Genesis) ChainID() string {
	return fmt.Sprintf("%s-%X", gen.data.ChainName, gen.ShortHash())
}

func (gen *Genesis) GenesisTime() time.Time {
	return gen.data.GenesisTime
}

func (gen *Genesis) GlobalAccount() *account.Account {
	gAcc, _ := account.NewAccount(crypto.GlobalAddress)
	gAcc.SetBalance(gen.data.GlobalAccount.Balance)
	gAcc.SetCode(gen.data.GlobalAccount.Code)
	gAcc.SetPermissions(gen.data.GlobalAccount.Permissions)

	return gAcc
}

func (gen *Genesis) Accounts() []*account.Account {
	accs := make([]*account.Account, 0)

	/// add global account
	acc := gen.GlobalAccount()
	accs = append(accs, acc)

	/// add genesis accounts
	for _, genAcc := range gen.data.Accounts {
		acc, err := account.NewAccount(genAcc.Address)
		if err == nil {
			acc.SetPermissions(genAcc.Permissions)
			acc.AddToBalance(genAcc.Balance)
			accs = append(accs, acc)
		}
	}

	/// add genesis contracts
	for _, genCt := range gen.data.Contracts {
		acc, err := account.NewAccount(genCt.Address)
		if err == nil {
			acc.SetPermissions(genCt.Permissions)
			acc.SetCode(genCt.Code)
			accs = append(accs, acc)
		}
	}

	return accs
}

func (gen *Genesis) Validators() []*validator.Validator {
	vals := make([]*validator.Validator, 0, len(gen.data.Validators))
	for _, genVal := range gen.data.Validators {
		val, err := validator.NewValidator(genVal.PublicKey, 0)
		if err == nil {
			val.AddToStake(genVal.Stake)
			vals = append(vals, val)
		}
	}

	return vals
}

func (gen *Genesis) MaximumPower() int {
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

func makeGenesisAccount(acc *account.Account) genAccount {
	return genAccount{
		Address:     acc.Address(),
		Balance:     acc.Balance(),
		Permissions: acc.Permissions(),
	}
}

func makeGenesisContract(acc *account.Account) genContract {
	return genContract{
		Address:     acc.Address(),
		Code:        acc.Code(),
		Permissions: acc.Permissions(),
	}
}

func makeGenesisValidator(val *validator.Validator) genValidator {
	return genValidator{
		PublicKey: val.PublicKey(),
		Address:   val.Address(),
		Stake:     val.Stake(),
	}
}

// MakeGenesis takes a chainName and a slice of pointers to Account,
// and a slice of pointers to Validator to construct a Genesis, or returns an error on
// failure.
func MakeGenesis(chainName string, genesisTime time.Time,
	globAccount *account.Account,
	accounts []*account.Account,
	contracts []*account.Account,
	validators []*validator.Validator) *Genesis {

	// Establish deterministic order of accs by address so we obtain identical Genesis
	// from identical input
	sort.SliceStable(accounts, func(i, j int) bool {
		return bytes.Compare(accounts[i].Address().RawBytes(), accounts[j].Address().RawBytes()) < 0
	})

	sort.SliceStable(contracts, func(i, j int) bool {
		return bytes.Compare(contracts[i].Address().RawBytes(), contracts[j].Address().RawBytes()) < 0
	})

	sort.SliceStable(validators, func(i, j int) bool {
		return bytes.Compare(validators[i].Address().RawBytes(), validators[j].Address().RawBytes()) < 0
	})

	genAccs := make([]genAccount, 0, len(accounts))
	for _, acc := range accounts {
		genAcc := makeGenesisAccount(acc)
		genAccs = append(genAccs, genAcc)
	}

	genCts := make([]genContract, 0, len(contracts))
	for _, ct := range contracts {
		genCt := makeGenesisContract(ct)
		genCts = append(genCts, genCt)
	}

	genVals := make([]genValidator, 0, len(validators))
	for _, val := range validators {
		genVal := makeGenesisValidator(val)
		genVals = append(genVals, genVal)
	}
	gAcc := globalAccount{
		Balance:     globAccount.Balance(),
		Code:        globAccount.Code(),
		Permissions: globAccount.Permissions(),
	}

	return &Genesis{
		data: genesisData{
			ChainName:     chainName,
			GenesisTime:   genesisTime,
			GlobalAccount: gAcc,
			Accounts:      genAccs,
			Contracts:     genCts,
			Validators:    genVals,
		},
	}
}

// LoadFromFile loads genesis object from a JSON file
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


/* save genesis file to file system */
func (gen *Genesis) Save(workingDir string) string {

	wDir := " "
	if workingDir != "" {
		wDir = workingDir
	} else {
		wDir = "/tmp/chain/"
	}
	filedir := wDir + "genesis.json"
	gene, generr := gen.MarshalJSON()
	if (generr) != nil {
		log.Fatalf("genesis error %s", generr)
	}
	/* create the directory*/
	if err := os.MkdirAll(filepath.Dir(filedir), 0777); err != nil {
		log.Fatalf("could not create directory %s", filepath.Dir(filedir))
	}
	/* write  dataContent to file */
	if err := ioutil.WriteFile(filedir, gene, 0600); err != nil {
		log.Fatalf("failed to write genesisfile to %s: %v", filedir, err)
	}

	msg := "created at " + filedir
	return msg

}
