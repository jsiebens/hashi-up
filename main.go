package main

import (
	"github.com/jsiebens/hashi-up/cmd"
	"os"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
