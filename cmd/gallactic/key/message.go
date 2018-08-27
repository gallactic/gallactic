package key

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gallactic/gallactic/crypto"
	"github.com/gallactic/gallactic/keystore/key"
	"github.com/jawher/mow.cli"
)

//Sign signs the message with the private key and returns the signature hash
func Sign() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		messageFilePath := cmd.String(cli.StringOpt{
			Name: "f file",
			Desc: "message file path to read the file and sign the message inside",
		})
		message := cmd.String(cli.StringOpt{
			Name: "m message",
			Desc: "message to sign",
		})
		privatekey := cmd.String(cli.StringOpt{
			Name: "p privatekey",
			Desc: "private key of the account to sign the message",
		})
		keyfilePath := cmd.String(cli.StringOpt{
			Name: "k keyfilePath",
			Desc: "path to the key file",
		})
		keyfileauthOpt := cmd.String(cli.StringOpt{
			Name: "a auth",
			Desc: "keyfile password",
		})

		cmd.Spec = "[--file=<messageFilePath>] | [--message=<message to sign>]" +
			" [--privatekey=<private key the account>] | [--keyfilePath=<path to the keyfile>] [--auth=<keyfile password>]"

		cmd.Action = func() {
			var msg []byte
			var err error
			//extract the message to be signed
			if *message != "" {
				msg = []byte(*message)
			} else if *messageFilePath != "" {
				msg, err = ioutil.ReadFile(*messageFilePath)
				if err != nil {
					log.Fatalf("Can't read message file: %v", err)
				}
			}
			var signature crypto.Signature
			var pv crypto.PrivateKey
			//Sign the message with the private key
			if *privatekey != "" {
				pv, err = crypto.PrivateKeyFromString(*privatekey)
				if err != nil {
					log.Fatalf("Could not obtain privatekey: %v", err)
				}
				signature, err = pv.Sign(msg)
				if err != nil {
					log.Fatalf("Error in signing: %v", err)
				}
			} else if *keyfilePath != "" {
				var passphrase string
				if *keyfileauthOpt == "" {
					passphrase = PromptPassphrase(true)
				} else {
					passphrase = *keyfileauthOpt
				}

				kj, err := key.DecryptKeyFile(*keyfilePath, passphrase)
				if err != nil {
					log.Fatalf("Could not decrypt file: %v", err)
				}
				pv = kj.PrivateKey()
				signature, err = pv.Sign(msg)
			}
			//display the signature hash
			fmt.Println(signature.String())
		}
	}
}

//Verify the signature of the signed message
func Verify() func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		publicKeyOpt := cmd.String(cli.StringArg{
			Name: "PUBLICKEY",
			Desc: "public key of the account",
		})
		signature := cmd.String(cli.StringArg{
			Name: "SIGNATURE",
			Desc: "Signature of the message",
		})
		message := cmd.String(cli.StringOpt{
			Name: "m message",
			Desc: "Message to be verified",
		})
		messageFilePath := cmd.String(cli.StringOpt{
			Name: "f messagefile",
			Desc: "Message File to be verified",
		})

		cmd.Spec = "PUBLICKEY SIGNATURE [--message=<message to be verified>] | [--messagefile=<Message File to be verified>]"
		cmd.Action = func() {
			var msg []byte
			var err error
			if *messageFilePath != "" {
				msg, err = ioutil.ReadFile(*messageFilePath)
				if err != nil {
					log.Fatalf("Error in reading File: %v", err)
				}
			} else {
				msg = []byte(*message)
			}
			var sign crypto.Signature
			publickey, err := crypto.PublicKeyFromString(*publicKeyOpt)
			if err != nil {
				log.Fatalf("Invalid public key %v", err)
			}
			sign, err = crypto.SignatureFromString(*signature)
			if err != nil {
				log.Fatalf("Invalid signature %v", err)
			}
			verify := publickey.Verify(msg, sign)
			if verify {
				fmt.Println("Signature Verification successfull!")
			} else {
				fmt.Println("Signature Verification failed!")
			}
		}
	}
}
