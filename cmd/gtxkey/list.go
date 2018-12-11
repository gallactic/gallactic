package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/gallactic/gallactic/keystore"
)

func List() func(c *ishell.Context) {
	return func(c *ishell.Context) {
		// open keystore
		ks := c.Get("keystore").(*keystore.Keystore)

		maxSpacing := len(strconv.Itoa(len(ks.Keys))) + 2
		spacing := generateChar(maxSpacing-2, " ")
		border := generateChar(50, "—")

		// print to console the header of the list
		fmt.Printf("\nNo.%s| status | Label (Address)", spacing)
		fmt.Printf("\n%s\n", border)

		// print to console the list of keys with certain infos
		for i, key := range ks.Keys {
			spacing = generateChar(maxSpacing-len(strconv.Itoa(i+1)), " ")
			status := "X"
			if key.Key != nil {
				status = "O"
			}

			fmt.Printf("%d.%s|   %s    | %s (%s) \n", i+1, spacing, status, key.Label, key.Address)
		}

		// print to console the legend for the status field
		fmt.Printf("\nNB: X - Locked, O - Unlocked\n\n")
	}
}

func generateChar(n int, s string) string {
	return strings.Repeat(s, n)
}
