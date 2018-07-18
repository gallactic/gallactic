package e

import (
	"fmt"
)

const (
	ErrNone                     = 0
	ErrGeneric                  = 1
	ErrTimeOut                  = 2
	ErrMultipleErrors           = 3
	ErrLoadingState             = 4
	ErrSavingState              = 5
	ErrInvalidAddress           = 500
	ErrInvalidPublicKey         = 501
	ErrInvalidPrivateKey        = 502
	ErrInvalidSignature         = 503
	ErrVMUnknownAddress         = 1000
	ErrVMInsufficientBalance    = 1001
	ErrVMInvalidJumpDest        = 1002
	ErrVMInsufficientGas        = 1003
	ErrVMMemoryOutOfBounds      = 1004
	ErrVMCodeOutOfBounds        = 1005
	ErrVMInputOutOfBounds       = 1006
	ErrVMReturnDataOutOfBounds  = 1007
	ErrVMCallStackOverflow      = 1008
	ErrVMCallStackUnderflow     = 1009
	ErrVMDataStackOverflow      = 1010
	ErrVMDataStackUnderflow     = 1011
	ErrVMInvalidContract        = 1012
	ErrVMNativeContractCodeCopy = 1013
	ErrVMExecutionAborted       = 1014
	ErrVMExecutionReverted      = 1015
	ErrVMNativeFunction         = 1016
	ErrVMNestedCall             = 1017
	ErrVMTransferValue          = 1018
	ErrVMEventPublish           = 1019
	ErrTxInvalidAddress         = 2000
	ErrTxDuplicateAddress       = 2001
	ErrInvalidAmount            = 2002
	ErrInsufficientFunds        = 2003
	ErrInvalidSequence          = 2007
	ErrTxWrongPayload           = 2008
	ErrPermInvalid              = 3000
	ErrPermDenied               = 3001
	ErrValChanged               = 4000
)

var messages = map[int]string{
	ErrNone:                     "No error",
	ErrGeneric:                  "Generic error",
	ErrTimeOut:                  "Timeout error",
	ErrMultipleErrors:           "Multiple errors",
	ErrInvalidAddress:           "Invalid address",
	ErrInvalidPublicKey:         "Invalid public key",
	ErrInvalidPrivateKey:        "Invalid private key",
	ErrInvalidSignature:         "Invalid signature",
	ErrVMUnknownAddress:         "Unknown address",
	ErrVMInsufficientBalance:    "Insufficient balance",
	ErrVMInvalidJumpDest:        "Invalid jump dest",
	ErrVMInsufficientGas:        "Insufficient gas",
	ErrVMMemoryOutOfBounds:      "Memory out of bounds",
	ErrVMCodeOutOfBounds:        "Code out of bounds",
	ErrVMInputOutOfBounds:       "Input out of bounds",
	ErrVMReturnDataOutOfBounds:  "Return data out of bounds",
	ErrVMCallStackOverflow:      "Call stack overflow",
	ErrVMCallStackUnderflow:     "Call stack underflow",
	ErrVMDataStackOverflow:      "Data stack overflow",
	ErrVMDataStackUnderflow:     "Data stack underflow",
	ErrVMInvalidContract:        "Invalid contract",
	ErrVMNativeContractCodeCopy: "Tried to copy native contract code",
	ErrVMExecutionAborted:       "Execution aborted",
	ErrVMExecutionReverted:      "Execution reverted",
	ErrVMNativeFunction:         "Native function error",
	ErrVMNestedCall:             "Error in nested call",
	ErrVMEventPublish:           "Event publish error",
	ErrVMTransferValue:          "Error transferring value ",
	ErrTxInvalidAddress:         "Invalid address",
	ErrTxDuplicateAddress:       "error duplicate address",
	ErrInvalidAmount:            "error invalid amount",
	ErrInsufficientFunds:        "error insufficient funds",
	ErrInvalidSequence:          "Error invalid sequence",
	ErrTxWrongPayload:           "Wrong payload",
	ErrPermInvalid:              "Invalid permission",
	ErrPermDenied:               "Permission denied",
	ErrValChanged:               "Validator has changed",
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
