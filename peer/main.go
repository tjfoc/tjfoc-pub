package main

import (
	"os"

	"github.com/tjfoc/tjfoc/peer/cmd"
)

func main() {
	if cmd.RootCmd.Execute() != nil {
		os.Exit(1)
	}
}
