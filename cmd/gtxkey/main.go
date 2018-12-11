package main

import (
	"github.com/abiosoft/ishell"
	"github.com/gallactic/gallactic/common"
	"github.com/gallactic/gallactic/keystore"
	"github.com/gallactic/gallactic/version"
)

var title = `
          __          __
    _____/  |____  __|  | __ ____ ___.__.
   / ___\   __\  \/  /  |/ // __ <   |  |
  / /_/  >  |  >    <|    <\  ___/\___  |
  \___  /|__| /__/\_ \__|_ \\___  > ____|
 /_____/            \/    \/    \/\/
`

func main() {
	shell := ishell.New()

	path := common.GallacticKeystoreDir()
	ks := keystore.Open(path)

	shell.Set("path", path)
	shell.Set("keystore", ks)

	shell.Println(title)
	shell.Println("Gallactic keystore version " + version.Version)
	shell.Println("Type `help` to start...")

	shell.AddCmd(&ishell.Cmd{
		Name: "new",
		Help: "Create new account",
		Func: New(),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "list",
		Help: "List all existing account",
		Func: List(),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "unlock",
		Help: "Unlock an account",
		Func: Unlock(),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "lock",
		Help: "Lock an existing account",
		Func: Lock(),
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "delete",
		Help: "Delete an account",
		Func: Delete(),
	})

	// run shell
	shell.Run()
}
