package crypto

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobalAddress(t *testing.T) {
	addr := "0000000000000000000000000000000000000000"
	bs, _ := hex.DecodeString(addr)
	gb, err := addressFromHash(bs, prefixGlobalAddress)
	fmt.Println(gb.String())

	assert.NoError(t, err)
	assert.Equal(t, GlobalAddress, gb)
}

/// http://lenschulwitz.com/base58
// 0590000000000000000000000000000000000000000000000000000000000000000000000000
// 0590ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff

// 12EC000000000000000000000000000000000000000000000000 - ac8CpYzraGsdAWaNdZESoxMTqkX2JmSeF7M
// 12ECffffffffffffffffffffffffffffffffffffffffffffffff - acXYRY79rz463L1Wiaen8SUjdP2HFUgE5Rg

// 1E2A000000000000000000000000000000000000000000000000 - vaAvnKWhjg5EnC1dwzf43ZCp18JX6qTTjPm
// 1E2Affffffffffffffffffffffffffffffffffffffffffffffff - vaaGPJd12PFhf1Sn325PN3L5nkon3Yh3Zi6

// 1434000000000000000000000000000000000000000000000000 - ct74fWcd5A3cYkk73aqhM1UjtByXhxFyUrT
// 1434ffffffffffffffffffffffffffffffffffffffffffffffff - ctWQGVivMsE5RaBF8cG2fVc1fpUnefVZKAn

// 13E90000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 - skGBELXCNmU421vrhfSYeaGFk7Xpt7Mii5KYdYsNWjZf8GJobnziogEzPLC8RbD6yogzywygtaZhWJxRh2y8Z2ShpwN92pf
// 13E9ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff - skqfmiwcQorVqaiycr4m3fpiE6HxhQtBkhb17imvwy7Y9TgyUYKByZee1vzcdofXdRd93E3Zw7frGy7bcqE38HycfmJ12wG

// 0590000000000000000000000000000000000000000000000000000000000000000000000000 - pj54DC8UUr5V1CXKvSqsDemm3RwahkyuX9rJQJXYY8XFiYYxuaF
// 0590777777777777777777777777777777777777777777777777777777777777777777777777 - pjyfpuCtYaXXGSzf1efbjLPKbx6WwVEytT9gee3mCdrmYhegiaJ
// 0590DDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD - pkkmVnz6tMCQvNpwNgWoK4dENPwbzyBLChyj9VvXCVrvYYsshS4
// 0590ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff - pm1oPRaWLGki9272q2TsAyNXxYZJMTpnJnvQeTDSXnXeDAcwMie

func TestCheckPrefix(t *testing.T) {
	for i := 0; i < 100; i++ {
		pb, pv := GenerateKey(nil)
		assert.True(t, strings.HasPrefix(pv.String(), "sk"))
		//assert.True(t, strings.HasPrefix(pb.String(), "pk"))
		assert.True(t, strings.HasPrefix(pb.AccountAddress().String(), "ac"))
		assert.True(t, strings.HasPrefix(pb.ValidatorAddress().String(), "va"))
		assert.True(t, strings.HasPrefix(DeriveContractAddress(pb.AccountAddress(), uint64(i)).String(), "ct"))
	}
}
