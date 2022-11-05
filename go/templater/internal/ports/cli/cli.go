package cli

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	templater "github.com/kendru/darwin/go/templater/internal"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "templater <template-dir> [output-dir]",
		Short: "Project Template Renderer",
		Long: `An application for rendering project templates to a directory.

Reads a template directory, template-dir, containing a manifest file, prompts
for input, and renders the result to output-dir. Output directory defaults to
the current directory.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inDir := os.Args[1]
			if info, err := os.Stat(inDir); err == nil {
				if !info.IsDir() {
					log.Fatalf("path is not a directory: %s\n", inDir)
				}
			} else {
				log.Fatalf("could not read directory: %v\n", err)
			}

			var outDir string
			if len(os.Args) == 3 {
				outDir = os.Args[2]
			} else {
				var err error
				if outDir, err = os.Getwd(); err != nil {
					return err
				}
			}

			tpl, err := templater.FromManifest(path.Join(inDir, "manifest.hcl"))
			if err != nil {
				return err
			}

			// Get variable input.
			inReader := bufio.NewReader(os.Stdin)
			inputs := tpl.GetInputs()
			vars := make(map[string]interface{}, len(inputs))
			for _, v := range inputs {
				// TODO: Allow config to be passed to CLI. Only prompt if missing inputs.
				var prompt strings.Builder

				prompt.WriteString(v.Name)
				if v.Description != nil {
					prompt.WriteString(" - ")
					prompt.WriteString(*v.Description)
				}
				if v.DefaultValue != nil {
					fmt.Fprintf(&prompt, " (%s)", v.DefaultValue)
				}
				prompt.WriteString(": ")
				fmt.Print(prompt.String())

				in, err := inReader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("reading input: %w", err)
				}
				in = strings.Trim(in, " \t\n\r")
				if len(in) == 0 {
					continue
				}

				if vars[v.Name], err = v.Coerce(in); err != nil {
					return err
				}
			}

			return tpl.Render(vars, outDir)
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
