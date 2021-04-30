package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	saveCmd := &cobra.Command{
		Use:   "save <name>",
		Short: "Saves the last command",
		Long: `Saves the last command run as a snip.
The command will be executed using your default shell.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if jsonBytes, err := json.MarshalIndent(args, "", "  "); err == nil {
				fmt.Println(string(jsonBytes))
			} else {
				fmt.Println("Error printing args:", err)
			}
		},
	}

	rootCmd.AddCommand(saveCmd)
}
