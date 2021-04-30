package jsondb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kendru/darwin/go/snip/internal"
)

type JSONDB struct {
	path string
}

func NewJSONDB(path string) *JSONDB {
	return &JSONDB{
		path: path,
	}
}

type serializableCommand struct {
	Interpreter string                  `json:"interpreter"`
	Text        string                  `json:"text"`
	Parameters  []serializableParameter `json:"parameters"`
}

type serializableParameter struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Default  interface{} `json:"default"`
	Required bool        `json:"required"`
}

func (db *JSONDB) Find(name string) (*internal.Command, error) {
	data, err := readDB(db.path)
	if err != nil {
		return nil, err
	}

	cmd, ok := data[name]
	if !ok {
		return nil, fmt.Errorf("snip not found: %q", name)
	}

	parameters := make([]internal.TemplateParameter, len(cmd.Parameters))
	for i, parameter := range cmd.Parameters {
		parameters[i] = internal.TemplateParameter{
			Name: parameter.Name,
			// TODO: validate type.
			Type:     parameter.Type,
			Default:  parameter.Default,
			Required: parameter.Required,
		}
	}

	return &internal.Command{
		Name:        name,
		Interpreter: cmd.Interpreter,
		Text:        cmd.Text,
		Parameters:  parameters,
	}, nil
}

func readDB(path string) (map[string]serializableCommand, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return createDB(path)
		}
		return nil, fmt.Errorf("opening database file: %w", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("reading database file: %w", err)
	}

	var data map[string]serializableCommand
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("decoding database file: %w", err)
	}

	return data, nil
}

func createDB(path string) (map[string]serializableCommand, error) {
	// TODO: Try to make directory as well.
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("creating database file: %w", err)
	}
	defer file.Close()

	if _, err = file.WriteString("{}"); err != nil {
		return nil, fmt.Errorf("creating database file: %w", err)
	}

	return make(map[string]serializableCommand), nil
}
