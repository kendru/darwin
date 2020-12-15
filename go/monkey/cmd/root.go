package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "monkey",
	Short: "Monkey is a simple programming language",
	Long:  `A programming languaged based on the book "Writing an Interpreter in Go"`,
	//Run: func(cmd *cobra.Command, args []string) {
	// Do Stuff Here
	//},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
