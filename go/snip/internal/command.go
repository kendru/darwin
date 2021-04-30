package internal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"text/template"
)

type Command struct {
	Name        string
	Interpreter string
	Text        string
	Parameters  []TemplateParameter
}

type TemplateParameter struct {
	Name     string
	Type     string
	Default  interface{}
	Required bool
}

func (cmd *Command) Run(ctx context.Context, args map[string]interface{}) error {
	interp, err := FindInterp(ctx, cmd.Interpreter)
	if err != nil {
		return err
	}

	text, err := cmd.commandString(ctx, args)
	if err != nil {
		return err
	}

	return interp.Run(ctx, text)
}

func (cmd *Command) commandString(ctx context.Context, args map[string]interface{}) (string, error) {
	var err error
	tpl := template.New(cmd.Name)
	if tpl, err = tpl.Parse(cmd.Text); err != nil {
		return "", fmt.Errorf("parsing command as template: %w", err)
	}

	fmt.Printf("PARAMS: %v\n", cmd.Parameters)
	for _, param := range cmd.Parameters {
		key := param.Name
		// TODO: Validate type.
		if _, ok := args[key]; !ok {
			if param.Default == nil {
				return "", fmt.Errorf("missing required argument: %q", key)
			}
			fmt.Printf("Setting %q to %v\n", key, param.Default)
			args[key] = param.Default
		}
	}

	// TODO: Validate args.
	var out bytes.Buffer
	tpl.Execute(&out, args)

	return out.String(), nil
}

type Interpreter struct {
	path string
	args []string
}

func (interp *Interpreter) Run(ctx context.Context, text string) error {
	args := append(append([]string{}, interp.args...), text)

	cmd := exec.Command(interp.path, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func FindInterp(ctx context.Context, name string) (*Interpreter, error) {
	// TODO: Open dispatch; user-defined interpreters.
	switch name {
	case "bash":
		return &Interpreter{
			path: "/bin/bash",
			args: []string{"-c"},
		}, nil
	default:
		return nil, fmt.Errorf("unknown interpreter: %q", name)
	}
}
