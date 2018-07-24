package validator

import (
	"testing"

	"github.com/gallactic/gallactic/crypto"
	"github.com/stretchr/testify/assert"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// TODO: add test for json marshaling

func TestValidatorSet(t *testing.T) {
	publicKeys := generatePublickKeys()
	validators := make([]*Validator, 6)
	validators[0] = NewValidator(publicKeys[0], 100, 1)
	validators[1] = NewValidator(publicKeys[1], 200, 1)
	validators[2] = NewValidator(publicKeys[2], 300, 1)
	validators[3] = NewValidator(publicKeys[3], 400, 1)
	validators[4] = NewValidator(publicKeys[4], 500, 1)
	validators[5] = NewValidator(publicKeys[5], 600, 1)

	vs := NewValidatorSet(validators)

	val := NewValidator(publickKeyFromSecret("z"), 100, 1)

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
	var err error

	for i, p := range publicKeys {
		validators[i] = NewValidator(p, 1, 100)
	}

	vs := ValidatorSet{
		data: validatorSetData{
			MaximumPower: 8,
			Validators:   validators[0:4],
		},
		proxy: proxy,
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

func compareValidators(validators1 []*Validator, tmValidators []*tmTypes.Validator) bool {

	if len(validators1) != len(tmValidators) {
		return false
	}

	for _, v1 := range validators1 {
		found := false
		for _, v2 := range validators1 {
			if v1.Address().EqualsTo(v2.Address()) {
				found = true
				break
			}
		}
		if found == false {
			return false
		}
	}

	return true
}

func publickKeyFromSecret(secret string) crypto.PublicKey {
	return crypto.PrivateKeyFromSecret(secret).PublicKey()
}

func generatePublickKeys() []crypto.PublicKey {
	publicKey := make([]crypto.PublicKey, 26)

	/// sorted by address
	publicKey[0] = publickKeyFromSecret("m")  //  18A71D0D81CEEBF548019C4BC24BB6F5B4E1361F
	publicKey[1] = publickKeyFromSecret("w")  //  1B4557CC1850966A88DCF7094F18ACC6756F1250
	publicKey[2] = publickKeyFromSecret("c")  //  366FC725E46FFDE3E63152AEA34B6EA15816D47D
	publicKey[3] = publickKeyFromSecret("x")  //  3F56ED107D8A808AEB2AAB523B72F7C37C812894
	publicKey[4] = publickKeyFromSecret("v")  //  4203FC4AE98849F0A1B6CB7E027FDE2FABD7AC62
	publicKey[5] = publickKeyFromSecret("a")  //  433CA69C9F597C9CD105740B04FC8CBFF206B587
	publicKey[6] = publickKeyFromSecret("r")  //  4861B368170E44623B86359B234CD4C485205678
	publicKey[7] = publickKeyFromSecret("z")  //  494F9624293B91E23C3D2AD946BB020F79D73CA8
	publicKey[8] = publickKeyFromSecret("t")  //  61BE4158B77C63BF69C8AF6733614F67C7DB45BA
	publicKey[9] = publickKeyFromSecret("n")  //  644CD981E309F6230F71C6434164205C64F82463
	publicKey[10] = publickKeyFromSecret("k") //  695AB9E2D56F83EA8403C007F17CCCB37A398594
	publicKey[11] = publickKeyFromSecret("i") //  7A37293D9152D3BE4A61DC4E79E7357421F212CC
	publicKey[12] = publickKeyFromSecret("j") //  8426FF76304CEAF18EE662B75E1B277303CD498C
	publicKey[13] = publickKeyFromSecret("d") //  8DFB3FDB0F0852D11BD58C231F09CEE35B78A376
	publicKey[14] = publickKeyFromSecret("q") //  91B3B57CA5921AC9F31A359F709A517F1D37A709
	publicKey[15] = publickKeyFromSecret("h") //  9CB5809A3FC3E9201C9039123F4BA39BD87F76FD
	publicKey[16] = publickKeyFromSecret("b") //  9E9AC3380A4941075BD2DDD534D7524BDDA6BB15
	publicKey[17] = publickKeyFromSecret("s") //  A49B77DB290F20764C70B46C16FB5D6801F70362
	publicKey[18] = publickKeyFromSecret("u") //  A83D2DABD3477EB60DA9A25343173BE7B4454728
	publicKey[19] = publickKeyFromSecret("y") //  AA6357CBBA4CF942178D03A02AC258BE168B7BCE
	publicKey[20] = publickKeyFromSecret("p") //  B044193ACBE2144DE94A7DA85E6A86DFF359C85E
	publicKey[21] = publickKeyFromSecret("l") //  C1461B8B1DC8D838A44B44E6ED423708258508DB
	publicKey[22] = publickKeyFromSecret("e") //  C75F161831E073F28FDA2C8C6DE20DFBFE277CA1
	publicKey[23] = publickKeyFromSecret("f") //  CC7B4572884B5C99D546FB8170A2A650DDCBA78E
	publicKey[24] = publickKeyFromSecret("o") //  E68964E9A5DFCE79A04D7E1B5AEEC5795C94BC73
	publicKey[25] = publickKeyFromSecret("g") //  FA987DA7B094D32392AF377A5079FC1D30DCC214

	return publicKey
}
