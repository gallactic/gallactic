package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/peterh/liner"
)

type terminalPrompter struct {
	*liner.State
	warned     bool
	supported  bool
	normalMode liner.ModeApplier
	rawMode    liner.ModeApplier
}

// Stdin holds the stdin line reader (also using stdout for printing prompts).
// Only this reader may be used for input because it keeps an internal buffer.
// var
var Stdin = newTerminalPrompter()

// newTerminalPrompter creates a liner based user input prompter working off the
// standard input and output streams.
func newTerminalPrompter() *terminalPrompter {
	p := new(terminalPrompter)
	// Get the original mode before calling NewLiner.
	// This is usually regular "cooked" mode where characters echo.
	normalMode, _ := liner.TerminalMode()
	// Turn on liner. It switches to raw mode.
	p.State = liner.NewLiner()
	rawMode, err := liner.TerminalMode()
	if err != nil || !liner.TerminalSupported() {
		p.supported = false
	} else {
		p.supported = true
		p.normalMode = normalMode
		p.rawMode = rawMode
		// Switch back to normal mode while we're not prompting.
		normalMode.ApplyMode()
	}
	p.SetCtrlCAborts(true)
	p.SetTabCompletionStyle(liner.TabPrints)
	p.SetMultiLineMode(true)
	return p
}

// PromptPassword displays the given prompt to the user and requests some textual
// data to be entered, but one which must not be echoed out into the terminal.
// The method returns the input provided by the user.
func (p *terminalPrompter) PromptPassword(prompt string) (string, error) {
	if p.supported {
		p.rawMode.ApplyMode()
		defer p.normalMode.ApplyMode()
		return p.State.PasswordPrompt(prompt)
	}
	if !p.warned {
		fmt.Println("!! Unsupported terminal, password will be echoed.")
		p.warned = true
	}
	// Just as in Prompt, handle printing the prompt here instead of relying on liner.
	fmt.Print(prompt)
	pass, err := p.State.Prompt("")
	fmt.Println()
	return pass, err
}

// PromptInput displays the given prompt to the user and requests some textual
// data to be entered, returning the input of the user.
func (p *terminalPrompter) PromptInput(prompt string) (string, error) {
	if p.supported {
		p.rawMode.ApplyMode()
		defer p.normalMode.ApplyMode()
	} else {
		// liner tries to be smart about printing the prompt
		// and doesn't print anything if input is redirected.
		// Un-smart it by printing the prompt always.
		fmt.Print(prompt)
		prompt = ""
		defer fmt.Println()
	}
	return p.State.Prompt(prompt)
}

// PromptConfirm displays the given prompt to the user and requests a boolean
// choice to be made, returning that choice.
func (p *terminalPrompter) PromptConfirm(prompt string) (bool, error) {
	input, err := p.Prompt(prompt + " [y/N] ")
	if len(input) > 0 && strings.ToUpper(input[:1]) == "Y" {
		return true, nil
	}
	return false, err
}

func CreateKey(pv crypto.PrivateKey) *key.Key {
	addr := pv.PublicKey().ValidatorAddress()
	key, _ := key.NewKey(addr, pv)
	return key
}

// PromptPassphrase prompts the user for a passphrase. Set confirmation to true
// to require the user to confirm the passphrase.
func PromptPassphrase(prompt string, confirmation bool) string {
	passphrase, err := Stdin.PromptPassword(prompt)
	if err != nil {
		log.Fatalf("Failed to read passphrase: %v", err)
	}

	if confirmation {
		confirm, err := Stdin.PromptPassword("Repeat passphrase: ")
		if err != nil {
			log.Fatalf("Failed to read passphrase confirmation: %v", err)
		}
		if passphrase != confirm {
			log.Fatalf("Passphrases do not match")
		}
	}

	return passphrase
}

// Promptlabel prompts for an input string
func PromptInput(prompt string) string {
	input, err := Stdin.PromptInput(prompt)
	if err != nil {
		log.Fatalf("Failed to read input: %v", err)
	}
	return input
}

// PromptPrivateKey prompts the user to enter the private key,
// validates the private key, displays the validator address and
// starts the node after confirmation
func PromptPrivateKey(promp string, accountKey bool) (*key.Key, error) {
	privatekey, err := Stdin.PromptInput(promp)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Private Key %v", err)
	}
	pv, err := crypto.PrivateKeyFromString(privatekey)
	if err != nil {
		return nil, fmt.Errorf("This is not a valid private key: %v", err)
	}

	// Creat key object
	var addr crypto.Address
	if accountKey {
		addr = pv.PublicKey().AccountAddress()
	} else {
		addr = pv.PublicKey().ValidatorAddress()
	}
	key, _ := key.NewKey(addr, pv)

	return key, nil
	/*
		fmt.Println("The private key is assigned to validator address: ", (keyObj.Address().String()))
		fmt.Println("press 'y' to proceed or 'n' to exit")
		confirm, err := line.Prompt("y/n: ")
		if err != nil {
			return nil, fmt.Errorf("Failed to read confirmation: %v", err)
		}
		if confirm == "y" {
			log.Print("Running Blockchain")
			return keyObj, nil
		}
		return nil, fmt.Errorf("Abort")
	*/
}
