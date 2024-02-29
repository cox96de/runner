package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{}
	root.AddCommand(getInstallCommand())
	root.AddCommand(getRunCommand())
	err := root.Execute()
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}
