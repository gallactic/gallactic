// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package evm

import (
	"fmt"

	"strings"

	"github.com/gallactic/gallactic/common/binary"
	"github.com/gallactic/gallactic/core/account"
	"github.com/gallactic/gallactic/core/account/permission"
	"github.com/gallactic/gallactic/core/evm/abi"
	"github.com/gallactic/gallactic/core/evm/sha3"
	"github.com/gallactic/gallactic/core/state"
	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/errors"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
)

//
// SNative (from 'secure natives') are native (go) contracts that are dispatched
// based on account permissions and can access and modify an account's permissions
//

// Metadata for SNative contract. Acts as a call target from the EVM. Can be
// used to generate bindings in a smart contract languages.
type SNativeContractDescription struct {
	// Comment describing purpose of SNative contract and reason for assembling
	// the particular functions
	Comment string
	// Name of the SNative contract
	Name          string
	functionsByID map[abi.FunctionSelector]*SNativeFunctionDescription
	functions     []*SNativeFunctionDescription
}

// Metadata for SNative functions. Act as call targets for the EVM when
// collected into an SNativeContractDescription. Can be used to generate
// bindings in a smart contract languages.
type SNativeFunctionDescription struct {
	// Comment describing function's purpose, parameters, and return value
	Comment string
	// Function name (used to form signature)
	Name string
	// Function arguments (used to form signature)
	Args []abi.Arg
	// Function return value
	Return abi.Return
	// Permissions required to call function
	Permissions account.Permissions
	// Native function to which calls will be dispatched when a containing
	// contract is called with a FunctionSelector matching this NativeContract
	F NativeContract
}

func registerSNativeContracts() {
	for _, contract := range SNativeContracts() {
		if !RegisterNativeContract(contract.Address().Word256(), contract.Dispatch) {
			panic(fmt.Errorf("could not register SNative contract %s because address %s already registered",
				contract.Address(), contract.Name))
		}
	}
}

// SNativeContracts returns a map of all SNative contracts defined indexed by name
func SNativeContracts() map[string]*SNativeContractDescription {
	permTypeName := abi.Uint64TypeName
	contracts := []*SNativeContractDescription{
		NewSNativeContract(`
		* Interface for managing Secure Native authorizations.
		* @dev This interface describes the functions exposed by the SNative permissions layer in burrow.
		`,
			"Permissions",
			&SNativeFunctionDescription{`
			* @notice Sets the permission flags for an account. Makes them explicitly set (on or off).
			* @param _address account address
			* @param _permissions the permissions to set for the account
			* @return result the effective permissions flags on the account after the call
			`,
				"setPermissions",
				[]abi.Arg{
					abiArg("_address", abi.AddressTypeName),
					abiArg("_permissions", permTypeName),
				},
				abiReturn("result", permTypeName),
				permission.ModifyPermission,
				setPermissions},

			&SNativeFunctionDescription{`
			* @notice Unsets the permissions flags for an account. Causes permissions being unset to fall through to global permissions.
      		* @param _address account address
      		* @param _permissions the permissions flags to unset for the account
			* @return result the effective permissions flags on the account after the call
      `,
				"unsetPermissions",
				[]abi.Arg{
					abiArg("_address", abi.AddressTypeName),
					abiArg("_permissions", permTypeName)},
				abiReturn("result", permTypeName),
				permission.ModifyPermission,
				unsetPermissions},

			&SNativeFunctionDescription{`
			* @notice Indicates whether an account has a subset of permissions set
			* @param _address account address
			* @param _permissions the permissions flags (mask) to check whether enabled against permissions for the account
			* @return result whether account has the passed permissions flags set
			`,
				"hasPermissions",
				[]abi.Arg{
					abiArg("_address", abi.AddressTypeName),
					abiArg("_permissions", permTypeName)},
				abiReturn("result", abi.BoolTypeName),
				permission.ModifyPermission,
				hasPermissions},
		),
	}

	contractMap := make(map[string]*SNativeContractDescription, len(contracts))
	for _, contract := range contracts {
		if _, ok := contractMap[contract.Name]; ok {
			// If this happens we have a pseudo compile time error that will be caught
			// on native.go init()
			panic(fmt.Errorf("duplicate contract with name %s defined. "+
				"Contract names must be unique", contract.Name))
		}
		contractMap[contract.Name] = contract
	}
	return contractMap
}

// Create a new SNative contract description object by passing a comment, name
// and a list of member functions descriptions
func NewSNativeContract(comment, name string,
	functions ...*SNativeFunctionDescription) *SNativeContractDescription {

	functionsByID := make(map[abi.FunctionSelector]*SNativeFunctionDescription, len(functions))
	for _, f := range functions {
		fid := f.ID()
		otherF, ok := functionsByID[fid]
		if ok {
			panic(fmt.Errorf("function with ID %x already defined: %s", fid, otherF.Signature()))
		}
		functionsByID[fid] = f
	}
	return &SNativeContractDescription{
		Comment:       comment,
		Name:          name,
		functionsByID: functionsByID,
		functions:     functions,
	}
}

// This function is designed to be called from the EVM once a SNative contract
// has been selected. It is also placed in a registry by registerSNativeContracts
// So it can be looked up by SNative address
func (contract *SNativeContractDescription) Dispatch(st *state.State, caller *account.Account,
	args []byte, gas *uint64, logger *logging.Logger) (output []byte, err error) {

	logger = logger.With(structure.ScopeKey, "Dispatch", "contract_name", contract.Name)

	if len(args) < abi.FunctionSelectorLength {
		return nil, e.Errorf(e.ErrNativeFunction,
			"SNatives dispatch requires a 4-byte function identifier but arguments are only %v bytes long",
			len(args))
	}

	function, err := contract.FunctionByID(abi.FirstFourBytes(args))
	if err != nil {
		return nil, err
	}

	logger.TraceMsg("Dispatching to function",
		"caller", caller.Address(),
		"function_name", function.Name)

	remainingArgs := args[abi.FunctionSelectorLength:]

	// check if we have permission to call this function
	if !st.HasPermissions(caller, function.Permissions) {
		return nil, e.Errorf(e.ErrPermDenied, "account %s does not have SNative function call permission: %s", caller.Address(), function.Name)
	}

	// ensure there are enough arguments
	if len(remainingArgs) != function.NArgs()*binary.Word256Length {
		return nil, e.Errorf(e.ErrNativeFunction, "%s() takes %d arguments but got %d (with %d bytes unconsumed - should be 0)",
			function.Name, function.NArgs(), len(remainingArgs)/binary.Word256Length, len(remainingArgs)%binary.Word256Length)
	}

	// call the function
	return function.F(st, caller, remainingArgs, gas, logger)
}

// We define the address of an SNative contact as the last 20 bytes of the sha3
// hash of its name
func (contract *SNativeContractDescription) Address() crypto.Address {
	hash := sha3.Sha3([]byte(contract.Name))
	address, _ := crypto.ContractAddress(hash)
	return address
}

// Get function by calling identifier FunctionSelector
func (contract *SNativeContractDescription) FunctionByID(id abi.FunctionSelector) (*SNativeFunctionDescription, error) {
	f, ok := contract.functionsByID[id]
	if !ok {
		return nil,
			e.Errorf(e.ErrNativeFunction, "unknown SNative function with ID %x", id)
	}
	return f, nil
}

// Get function by name
func (contract *SNativeContractDescription) FunctionByName(name string) (*SNativeFunctionDescription, error) {
	for _, f := range contract.functions {
		if f.Name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("unknown SNative function with name %s", name)
}

// Get functions in order of declaration
func (contract *SNativeContractDescription) Functions() []*SNativeFunctionDescription {
	functions := make([]*SNativeFunctionDescription, len(contract.functions))
	copy(functions, contract.functions)
	return functions
}

//
// SNative functions
//

// Get function signature
func (function *SNativeFunctionDescription) Signature() string {
	argTypeNames := make([]string, len(function.Args))
	for i, arg := range function.Args {
		argTypeNames[i] = string(arg.TypeName)
	}
	return fmt.Sprintf("%s(%s)", function.Name,
		strings.Join(argTypeNames, ","))
}

// Get function calling identifier FunctionSelector
func (function *SNativeFunctionDescription) ID() abi.FunctionSelector {
	return abi.FunctionID(function.Signature())
}

// Get number of function arguments
func (function *SNativeFunctionDescription) NArgs() int {
	return len(function.Args)
}

func abiArg(name string, abiTypeName abi.TypeName) abi.Arg {
	return abi.Arg{
		Name:     name,
		TypeName: abiTypeName,
	}
}

func abiReturn(name string, abiTypeName abi.TypeName) abi.Return {
	return abi.Return{
		Name:     name,
		TypeName: abiTypeName,
	}
}

// Permission function defintions
func hasPermissions(st *state.State, caller *account.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, permNum := returnTwoArgs(args)
	addr, err := crypto.AddressFromWord256(addrWord256)
	if err != nil {
		return nil, err
	}

	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}

	perm := account.Permissions(binary.Uint64FromWord256(permNum))
	if !permission.EnsureValid(perm) {
		return nil, e.Error(e.ErrPermInvalid)
	}
	hasPermissions := st.HasPermissions(acc, perm)
	permInt := byteFromBool(hasPermissions)
	logger.Trace.Log("function", "hasPermissions",
		"address", addr.String(),
		"account.permissions", acc.Permissions().String(),
		"permissions", perm.String(),
		"has_permission", hasPermissions)
	return binary.LeftPadWord256([]byte{permInt}).Bytes(), nil
}

func setPermissions(st *state.State, caller *account.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, permNum := returnTwoArgs(args)
	addr, err := crypto.AddressFromWord256(addrWord256)
	if err != nil {
		return nil, err
	}

	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}

	perm := account.Permissions(binary.Uint64FromWord256(permNum))
	if !permission.EnsureValid(perm) {
		return nil, e.Error(e.ErrPermInvalid)
	}

	if err := acc.SetPermissions(perm); err != nil {
		return nil, err
	}
	st.UpdateAccount(acc)
	logger.Trace.Log("function", "setPermissions",
		"address", addr.String(),
		"account.permissions", acc.Permissions().String(),
		"permission", perm.String())
	return effectivePermBytes(acc.Permissions(), globalPerms(st)), nil
}

func unsetPermissions(st *state.State, caller *account.Account, args []byte, gas *uint64,
	logger *logging.Logger) (output []byte, err error) {

	addrWord256, permNum := returnTwoArgs(args)
	addr, err := crypto.AddressFromWord256(addrWord256)
	if err != nil {
		return nil, err
	}

	acc, err := st.GetAccount(addr)
	if err != nil {
		return nil, err
	}

	perm := account.Permissions(binary.Uint64FromWord256(permNum))
	if !permission.EnsureValid(perm) {
		return nil, e.Error(e.ErrPermInvalid)
	}

	if err = acc.UnsetPermissions(perm); err != nil {
		return nil, err
	}
	st.UpdateAccount(acc)
	logger.Trace.Log("function", "unsetPermissions",
		"address", addr.String(),
		"account.permissions", acc.Permissions().String(),
		"permission", perm.String())

	return effectivePermBytes(acc.Permissions(), globalPerms(st)), nil
}

//------------------------------------------------------------------------------------------------
// Errors and utility funcs

// Get the global BasePermissions
func globalPerms(st *state.State) account.Permissions {
	return st.GlobalAccount().Permissions()
}

// Compute the effective permissions from an acm.Account's BasePermissions by
// taking the bitwise or with the global BasePermissions resultant permissions
func effectivePermBytes(permissions, globalPermissions account.Permissions) []byte {
	return permBytes(permissions | globalPermissions)
}

func permBytes(perm account.Permissions) []byte {
	return binary.Uint64ToWord256(uint64(perm)).Bytes()
}

// CONTRACT: length has already been checked
func returnTwoArgs(args []byte) (a, b binary.Word256) {
	copy(a[:], args[:32])
	copy(b[:], args[32:64])
	return
}

// CONTRACT: length has already been checked
func returnThreeArgs(args []byte) (a, b, c binary.Word256) {
	copy(a[:], args[:32])
	copy(b[:], args[32:64])
	copy(c[:], args[64:96])
	return
}

func byteFromBool(b bool) byte {
	if b {
		return 0x1
	}
	return 0x0
}
