package main

import (
	"os"

	"github.com/bungogood/worktree/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
