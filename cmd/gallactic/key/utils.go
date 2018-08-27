package key

import (
	"fmt"
	"log"

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
var (
	Stdin = newTerminalPrompter()
)

// promptPassphrase prompts the user for a passphrase.  Set confirmation to true
// to require the user to confirm the passphrase.
func PromptPassphrase(confirmation bool) string {

	if confirmation {
		passphrase, err := Stdin.PromptPassword("New passphrase: ")
		if err != nil {
			log.Fatalf("Failed to read passphrase: %v", err)
		}

		confirm, err := Stdin.PromptPassword("Repeat passphrase: ")
		if err != nil {
			log.Fatalf("Failed to read passphrase confirmation: %v", err)
		}
		if passphrase != confirm {
			log.Fatalf("Passphrases do not match")
		}
		return passphrase
	}
	return "Error"
}

// PromptPassword displays the given prompt to the user and requests some textual
// data to be entered, but one which must not be echoed out into the terminal.
// The method returns the input provided by the user.
func (p *terminalPrompter) PromptPassword(prompt string) (passwd string, err error) {
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
	passwd, err = p.State.Prompt("")
	fmt.Println()
	return passwd, err
}

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

//PromptPrivateKey prompts the user to enter the private key,
// validates the private key, displays the validator address and
// starts the node after confirmation
func PromptPrivateKey() (*key.Key, error) {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	privatekey, err := line.Prompt("please enter your privateKey for validator: ")
	if err != nil {
		return nil, fmt.Errorf("Failed to read Private Key %v", err)
	} else if err == liner.ErrPromptAborted {
		log.Print("Aborted")
		return nil, err
	}
	pv, err := crypto.PrivateKeyFromString(privatekey)
	if err != nil {
		return nil, fmt.Errorf("This was not a valid private key ")
	}
	keyObj := CreateKey(pv)
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
}

// oldPassphrase prompts for the old password of the keyfile
func oldPassphrase() string {
	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	passphrase, err := line.Prompt("Old Password: ")
	if err != nil {
		fmt.Errorf("Failed to read passphrase: %v", err)
	}
	return passphrase
}

func CreateKey(pv crypto.PrivateKey) *key.Key {
	addr := pv.PublicKey().ValidatorAddress()
	return key.NewKey(addr, pv)
}
