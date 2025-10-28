package main

import (
	"os"

	"github.com/raucheacho/rosia-cli/cmd"
)

func main() {
	exitCode := cmd.ExecuteWithExitCode()
	os.Exit(exitCode)
}
