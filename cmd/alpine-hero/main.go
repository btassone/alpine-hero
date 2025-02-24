package main

import (
	"fmt"
	"os"

	"github.com/btassone/alpine-hero/cmd/alpine-hero/cmd"
)

func main() {
	if pErr := cmd.Execute(); pErr != nil {
		_, err := fmt.Fprintf(os.Stderr, "Error: %v\n", pErr)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
