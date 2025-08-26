package main

import (
	"os"

	"github.com/zxj777/claude-helper/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}