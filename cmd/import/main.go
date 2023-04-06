package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: import <cmd> where cmd is 'bladechain'|'users'")
	}
	cmd := args[1]

	switch cmd {
	case "bladechain":
		importBladechain("reference/bladechain2.txt")
	case "users":
		err := importUsers()
		if err != nil {
			log.Fatal(err)
		}
	}
}
