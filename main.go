package main

import (
	"fmt"
	"os"

	"github.com/bamboo-services/bamboo-base-go-cli/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
