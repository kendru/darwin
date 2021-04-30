package cmd

import (
	"context"
	"fmt"

	"github.com/kendru/darwin/go/snip/internal/jsondb"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	runCmd := &cobra.Command{
		Use:   "run <name>",
		Short: "Runs a command",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			db := jsondb.NewJSONDB("./data/db.json")
			snip, err := db.Find(name)
			if err != nil {
				return err
			}

			// TODO: Cannot easily use Cobra :(

			// TODO: Improve the interface
			for _, param := range snip.Parameters {
				// TODO: dispatch on type
				dflt := ""
				if param.Default != nil {
					dflt = param.Default.(string)
				}

				fmt.Printf("Setting param for %q\n", param.Name)
				cmd.Flags().String(param.Name, dflt, "TODO")
				viper.BindPFlag(param.Name, cmd.Flags().Lookup(param.Name))
			}

			invocationArgs := viper.AllSettings()

			return snip.Run(context.TODO(), invocationArgs)
		},
	}

	rootCmd.AddCommand(runCmd)
}
