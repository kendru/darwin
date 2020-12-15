package cmd

import (
	"fmt"
	"os"

	"github.com/kendru/darwin/go/monkey/repl"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(replCmd)
}

var replCmd = &cobra.Command{
	Use:   "repl",
	Short: "Start a Monkey REPL",
	Long:  `Start a Read-Eval-Print loop that evaluates code live`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting Monkey REPL.")
		repl.Start(os.Stdin, os.Stdout)
	},
}
