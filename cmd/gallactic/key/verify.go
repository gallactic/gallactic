package key

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gallactic/gallactic/crypto"
	"github.com/jawher/mow.cli"
)

//Verify the signature of the signed message
func Verify() func(c *cli.Cmd) {
	return func(c *cli.Cmd) {
		publicKey := c.String(cli.StringOpt{
			Name: "p publickey",
			Desc: "Public key of the account",
		})
		signature := c.String(cli.StringOpt{
			Name: "s signature",
			Desc: "Signature of the message",
		})
		message := c.String(cli.StringOpt{
			Name: "m message",
			Desc: "Message to be verified",
		})
		messageFile := c.String(cli.StringOpt{
			Name: "f messagefile",
			Desc: "Message File to be verified",
		})

		c.Spec = "[-p=<public key>] [-s=<signature>] [-m=<message to be verified>] | [-f=<Message File to be verified>]"
		c.Action = func() {
			var msg []byte
			var err error
			if *messageFile != "" {
				msg, err = ioutil.ReadFile(*messageFile)
				if err != nil {
					log.Fatalf("Error in reading File: %v", err)
				}
			} else {
				msg = []byte(*message)
			}
			var sign crypto.Signature
			publickey, err := crypto.PublicKeyFromString(*publicKey)
			if err != nil {
				log.Fatalf("Invalid public key %v", err)
			}
			sign, err = crypto.SignatureFromString(*signature)
			if err != nil {
				log.Fatalf("Invalid signature %v", err)
			}
			fmt.Println(string(msg))
			verify := publickey.Verify(msg, sign)
			if verify {
				fmt.Println("Signature Verification successfull!")
			} else {
				fmt.Println("Signature Verification failed!")
			}
		}
	}
}
