package validator

import (
	"bytes"
	"sort"
	"testing"

	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

func TestValidatorSet(t *testing.T) {
	publicKeys := generatePublickKeys()
	validators := make(map[crypto.Address]*Validator)
	validators[publicKeys[0].ValidatorAddress()], _ = NewValidator(publicKeys[0], 1)
	validators[publicKeys[1].ValidatorAddress()], _ = NewValidator(publicKeys[1], 1)
	validators[publicKeys[2].ValidatorAddress()], _ = NewValidator(publicKeys[2], 1)
	validators[publicKeys[3].ValidatorAddress()], _ = NewValidator(publicKeys[3], 1)
	validators[publicKeys[4].ValidatorAddress()], _ = NewValidator(publicKeys[4], 1)
	validators[publicKeys[5].ValidatorAddress()], _ = NewValidator(publicKeys[5], 1)

	vs := NewValidatorSet(validators, 8)

	val, _ := NewValidator(publickKeyFromSecret("z"), 1)

	err := vs.ForceLeave(val)
	assert.Error(t, err)
	assert.Equal(t, 6, vs.TotalPower())
	assert.Equal(t, false, vs.Contains(val.Address()))
	err = vs.Join(val)
	assert.NoError(t, err)
	assert.Equal(t, 7, vs.TotalPower())
	assert.Equal(t, true, vs.Contains(val.Address()))
	/// expecting an error, validator already exist in the set
	err = vs.Join(val)
	assert.Error(t, err)
	vs.ForceLeave(val)
	assert.Equal(t, 6, vs.TotalPower())
	assert.Equal(t, false, vs.Contains(val.Address()))
}

type _validatorListProxyMock struct {
	height        int64
	validatorSets [][]*tmTypes.Validator
}

func newValidatorListProxyMock() *_validatorListProxyMock {

	proxy := &_validatorListProxyMock{}
	publicKeys := generatePublickKeys()
	validators := make([]*tmTypes.Validator, len(publicKeys))

	for i, p := range publicKeys {
		tmPubKey := tmCrypto.PubKeyEd25519{}
		copy(tmPubKey[:], p.RawBytes())

		validators[i] = tmTypes.NewValidator(tmPubKey, 1)
	}

	/// round:1, power:4
	/// <- validator[0,1,2,3] joined
	proxy.nextRound(validators[0:4])

	/// round:2, power:5
	/// <- validator[4] joined
	proxy.nextRound(validators[0:5])

	/// round:3, power:6
	/// <- validator[5] joined
	proxy.nextRound(validators[0:6])

	/// round:4, power:7
	/// <- validator[6] joined
	proxy.nextRound(validators[0:7])

	/// round:5, power:8
	/// <- validator[7] joined
	proxy.nextRound(validators[0:8])

	/// round:6, power:8 (no change)
	proxy.nextRound(validators[0:8])

	/// round:7
	/// -> validator[0] left
	/// <- validator[8] joined
	proxy.nextRound(validators[1:9])

	/// round:8
	/// -> validator[1] left
	/// <- validator[9,10,11,12] joined
	proxy.nextRound(validators[2:13])

	/// round:9
	/// -> validator[2] left
	/// <- validator[13] joined
	proxy.nextRound(validators[3:14])

	/// round:10
	/// -> validator[3] left
	proxy.nextRound(validators[4:14])

	/// round:11
	/// -> validator[4] left
	proxy.nextRound(validators[5:14])

	/// round:12
	/// -> validator[5] left
	proxy.nextRound(validators[6:14])

	/// round:13
	/// -> validator[6] left
	proxy.nextRound(validators[6:14])

	/// round:14
	proxy.nextRound(validators[6:14])

	return proxy
}

func (proxy _validatorListProxyMock) validators(height int64) (*tmRPCTypes.ResultValidators, error) {
	var result tmRPCTypes.ResultValidators
	result.Validators = proxy.validatorSets[height-1]
	result.BlockHeight = height

	return &result, nil
}

func (proxy _validatorListProxyMock) tmValidators(height int64) []*tmTypes.Validator {
	result, _ := proxy.validators(height)

	return result.Validators
}

func (proxy *_validatorListProxyMock) nextRound(validators []*tmTypes.Validator) {
	tmValidators := make([]*tmTypes.Validator, len(validators))
	copy(tmValidators, validators)

	proxy.height++
	proxy.validatorSets = append(proxy.validatorSets, tmValidators)
}

func TestAdjusting(t *testing.T) {

	proxy := newValidatorListProxyMock()
	publicKeys := generatePublickKeys()
	validators := make([]*Validator, len(publicKeys))
	validatorMap := make(map[crypto.Address]*Validator)
	var err error

	for i, p := range publicKeys {
		validators[i], _ = NewValidator(p, 1)
	}

	for i := 0; i < 4; i++ {
		validatorMap[validators[i].Address()] = validators[i]
	}

	vs := ValidatorSet{
		maximumPower: 8,
		validators:   validatorMap,
		leavers:      make(map[crypto.Address]*Validator),
		proxy:        proxy,
	}

	// -----------------------------------------
	vs.Join(validators[4])
	err = vs.AdjustPower(2)

	assert.NoError(t, err)
	assert.Equal(t, 5, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(2)))

	// println(fmt.Sprintf("%v", vs.Validators()))
	// println(fmt.Sprintf("%v", proxy.tmValidators))

	// -----------------------------------------
	vs.Join(validators[5])
	err = vs.AdjustPower(3)

	assert.NoError(t, err)
	assert.Equal(t, 6, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(3)))

	// -----------------------------------------
	vs.Join(validators[6])
	err = vs.AdjustPower(4)

	assert.NoError(t, err)
	assert.Equal(t, 7, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(4)))

	// -----------------------------------------
	vs.Join(validators[7])
	err = vs.AdjustPower(5)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(5)))

	// -----------------------------------------
	err = vs.AdjustPower(6)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(6)))

	// -----------------------------------------
	vs.Join(validators[8])
	err = vs.AdjustPower(7)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(7)))

	// -----------------------------------------
	vs.Join(validators[9])
	vs.Join(validators[10])
	vs.Join(validators[11])
	vs.Join(validators[12])
	err = vs.AdjustPower(8)

	assert.NoError(t, err)
	assert.Equal(t, 11, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(8)))

	// -----------------------------------------
	vs.Join(validators[9])
	vs.Join(validators[10])
	vs.Join(validators[11])
	vs.Join(validators[12])
	vs.Join(validators[13])
	err = vs.AdjustPower(9)

	assert.NoError(t, err)
	assert.Equal(t, 11, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(9)))

	// -----------------------------------------
	err = vs.AdjustPower(10)

	assert.NoError(t, err)
	assert.Equal(t, 10, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(10)))

	// -----------------------------------------
	err = vs.AdjustPower(11)

	assert.NoError(t, err)
	assert.Equal(t, 9, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(11)))

	// -----------------------------------------
	err = vs.AdjustPower(12)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(12)))

	// -----------------------------------------
	err = vs.AdjustPower(13)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(13)))

	// -----------------------------------------
	err = vs.AdjustPower(14)

	assert.NoError(t, err)
	assert.Equal(t, 8, vs.TotalPower())
	assert.Equal(t, true, compareValidators(vs.Validators(), proxy.tmValidators(14)))
}

func compareValidators(validators1 map[crypto.Address]*Validator, tmValidators []*tmTypes.Validator) bool {

	if len(validators1) != len(tmValidators) {
		return false
	}

	for _, v := range tmValidators {
		addr, _ := crypto.ValidatorAddress(v.Address.Bytes())
		_, ok := validators1[addr]
		if !ok {
			return false
		}
	}

	return true
}

func publickKeyFromSecret(secret string) crypto.PublicKey {
	return crypto.PrivateKeyFromSecret(secret).PublicKey()
}

func generatePublickKeys() []crypto.PublicKey {
	publicKeys := make([]crypto.PublicKey, 26)

	/// sorted by address
	publicKeys[0] = publickKeyFromSecret("m")  // vaCwdxTmJQUkCMr85Zo7e43nqvEWigNqfek
	publicKeys[1] = publickKeyFromSecret("w")  // vaHtTqwBbwDyNWfFSKGYER7uGwFNg1iCimD
	publicKeys[2] = publickKeyFromSecret("c")  // vaVFZ28jjDyEAvjw2UtDQAttuCuiEKy24R6
	publicKeys[3] = publickKeyFromSecret("x")  // vaCTakoLDRMrZRZEvuPdC7xXehpkounNJAN
	publicKeys[4] = publickKeyFromSecret("v")  // vaYU6dNdPnAM5Q6CoPR3vqDGHXUJ95pEFAi
	publicKeys[5] = publickKeyFromSecret("a")  // vaEy3rRFBVt8yAxCzPAjr3qL2VJWdn3Q6LR
	publicKeys[6] = publickKeyFromSecret("r")  // vaV5D9ndeVSC8oGuGpXwMtRCS1ouaxim1P6
	publicKeys[7] = publickKeyFromSecret("z")  // vaBQvqznfToiDgKTFgiYgV7Q214ENCDuS63
	publicKeys[8] = publickKeyFromSecret("t")  // vaHQQymJ2fzQYFvRvZUDBybRzaDhMjF4MRJ
	publicKeys[9] = publickKeyFromSecret("n")  // vaCwr5Q2pGStDFiXTtrnXspW628xVeUvhBv
	publicKeys[10] = publickKeyFromSecret("k") // vaHPDBa14pHLv3vfyKnLDcq6oRmMv9vPJKS
	publicKeys[11] = publickKeyFromSecret("i") // vaCV3aa81M8fzjdxRfbsXSLu26Y3LA3dWti
	publicKeys[12] = publickKeyFromSecret("j") // vaBy3MyaCdXsWuULtuQ1HMDgCtnRxbRkMJ7
	publicKeys[13] = publickKeyFromSecret("d") // vaLFiiB1gMhC7ZJPaYR7ZJxbDnTjQaDes85
	publicKeys[14] = publickKeyFromSecret("q") // vaUfroLyxyMJ8pQqau5UeeSnZQxv5TLBTCS
	publicKeys[15] = publickKeyFromSecret("h") // vaXRchdAtC3nE8P9mcwaLc3RoPCNxGPc7nR
	publicKeys[16] = publickKeyFromSecret("b") // vaRGTNDViWGuxp9uUxcSsvwVRnAhDoCPubt
	publicKeys[17] = publickKeyFromSecret("s") // vaC4FTSA1GpTW651NcWyz88R2teyynxc8xH
	publicKeys[18] = publickKeyFromSecret("u") // vaMwP3ny4SoMsCyAVSYq5XMgaZMy8LEGLM2
	publicKeys[19] = publickKeyFromSecret("y") // vaVFmotUpFNxzYUZ3x4vRi6Z8i3tFo4Q7jY
	publicKeys[20] = publickKeyFromSecret("p") // vaJAreyU72KmXbLPw7k7rSY7LCBgD5KYF7Y
	publicKeys[21] = publickKeyFromSecret("l") // vaZgYD2XzXuXrQAE77EmEfGWwhe7P9JLuPA
	publicKeys[22] = publickKeyFromSecret("e") // vaT5rBh6UDKyxFFqPoJ4gqKcmhkkEL7uNDF
	publicKeys[23] = publickKeyFromSecret("f") // vaXw9oo51Za66j1YK8PLgDRcmGFuHjqN5Xw
	publicKeys[24] = publickKeyFromSecret("o") // vaZQCXRHd6Q8D62GxmyL7MzH41ZTd8rhMGr
	publicKeys[25] = publickKeyFromSecret("g") // vaJHdewyGjV4Zmaj1p92S4UZBRhg5MEmmbS

	// Sorting by address because of _validatorListProxyMock
	sort.SliceStable(publicKeys, func(i, j int) bool {
		return bytes.Compare(publicKeys[i].ValidatorAddress().RawBytes(), publicKeys[j].ValidatorAddress().RawBytes()) > 0
	})

	return publicKeys
}
