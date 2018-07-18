package codes

import (
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

const (
	// Success
	TxExecutionSuccessCode uint32 = abciTypes.CodeTypeOK

	// Informational
	UnsupportedRequestCode uint32 = 400

	// Internal errors
	EncodingErrorCode    uint32 = 500
	TxExecutionErrorCode uint32 = 501
	CommitErrorCode      uint32 = 502
)
