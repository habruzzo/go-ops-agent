package main

import (
	"os"

	"github.com/habruzzo/agent/cli"
)

func main() {
	app := cli.NewCLI()
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}
