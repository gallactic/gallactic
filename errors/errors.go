package e

import (
	"fmt"
)

const (
	ErrNone = iota
	ErrGeneric
	ErrTimeOut
	ErrInvalidAddress
	ErrInvalidPublicKey
	ErrInvalidPrivateKey
	ErrInvalidSignature
	ErrInvalidAmount
	ErrInvalidSequence
	ErrInvalidTxType
	ErrDuplicateAddress
	ErrInsufficientFunds
	ErrInsufficientGas
	ErrPermInvalid
	ErrPermDenied
	ErrNativeFunction
	ErrInternalEvm

	errCount
)

var messages = map[int]string{
	ErrNone:              "No error",
	ErrGeneric:           "Generic error",
	ErrTimeOut:           "Timeout error",
	ErrInvalidAddress:    "Invalid address",
	ErrInvalidPublicKey:  "Invalid public key",
	ErrInvalidPrivateKey: "Invalid private key",
	ErrInvalidSignature:  "Invalid signature",
	ErrInvalidAmount:     "error invalid amount",
	ErrInvalidSequence:   "Error invalid sequence",
	ErrInvalidTxType:     "Invalid transaction type",
	ErrDuplicateAddress:  "error duplicate address",
	ErrInsufficientFunds: "error insufficient funds",
	ErrInsufficientGas:   "Insufficient Gas",
	ErrPermInvalid:       "Invalid permission",
	ErrPermDenied:        "Permission denied",
	ErrNativeFunction:    "Error on calling native contracts",
	ErrInternalEvm:       "Internal EVM error",
}

type withCode struct {
	code    int
	message string
}

func Error(code int) error {
	message, ok := messages[code]
	if !ok {
		message = "Unknown error code"
	}

	return &withCode{
		code:    code,
		message: message,
	}
}

func Errorf(code int, format string, a ...interface{}) error {
	message, ok := messages[code]
	if !ok {
		message = "Unknown error code"
	}

	return &withCode{
		code:    code,
		message: message + ": " + fmt.Sprintf(format, a...),
	}
}

func Code(err error) int {
	type i interface {
		Code() int
	}

	if err == nil {
		return ErrNone
	}
	_e, ok := err.(i)
	if !ok {
		return ErrGeneric
	}
	return _e.Code()
}

func (e *withCode) Error() string {
	return e.message
}

func (e *withCode) Code() int {
	return e.code
}
