package e

import (
	"fmt"
)

const (
	ErrNone = iota
	ErrGeneric
	ErrTimeOut
	ErrMultipleErrors
	ErrInvalidAddress
	ErrInvalidPublicKey
	ErrInvalidPrivateKey
	ErrInvalidSignature
	ErrDuplicateAddress
	ErrInvalidAmount
	ErrInsufficientFunds
	ErrInvalidSequence
	ErrPermInvalid
	ErrPermDenied
	ErrNativeFunction
	ErrInsufficientGas
	ErrTxInvalidType

	errCount
)

var messages = map[int]string{
	ErrNone:              "No error",
	ErrGeneric:           "Generic error",
	ErrTimeOut:           "Timeout error",
	ErrMultipleErrors:    "Multiple errors",
	ErrInvalidAddress:    "Invalid address",
	ErrInvalidPublicKey:  "Invalid public key",
	ErrInvalidPrivateKey: "Invalid private key",
	ErrInvalidSignature:  "Invalid signature",
	ErrDuplicateAddress:  "error duplicate address",
	ErrInvalidAmount:     "error invalid amount",
	ErrInsufficientFunds: "error insufficient funds",
	ErrInvalidSequence:   "Error invalid sequence",
	ErrPermInvalid:       "Invalid permission",
	ErrPermDenied:        "Permission denied",
	ErrNativeFunction:    "Error on calling native contracts",
	ErrInsufficientGas:   "Insufficient Gas",
	ErrTxInvalidType:     "Invalid transaction type",
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

func Errors(errors ...error) error {
	message := messages[ErrMultipleErrors]

	for _, err := range errors {
		message += "\n"
		message += err.Error()
	}

	return &withCode{
		code:    ErrMultipleErrors,
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

func ErrorE(code int, err error) error {
	message, ok := messages[code]
	if !ok {
		message = "Unknown error code"
	}

	return &withCode{
		code:    code,
		message: message + ": " + err.Error(),
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
