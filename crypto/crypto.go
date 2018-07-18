package crypto

var GlobalAddress, _ = addressFromHash([]byte{0}, globalAddress)

func DeriveContractAddress(addr *Address, sequence uint64) Address {
	/*
		/// Ankur: Contract address has contract type
		temp := make([]byte, 32+8)
		copy(temp, addr.data.Address[:])
		binary.PutUint64BE(temp[32:], uint64(sequence))
		hasher := ripemd160.New()
		hasher.Write(temp) // does not error


		copy(newAddr[:], hasher.Sum(nil))
	*/
	return Address{}
}
