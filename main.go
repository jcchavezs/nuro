package main

import (
	"fmt"
	"os"

	"github.com/jcchavezs/nuro/internal/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		fmt.Printf("Failed to execute nuro: %v\n", err)
		os.Exit(1)
	}
}
