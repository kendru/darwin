package templater

import (
	"bytes"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/hcl/v2"
	"github.com/kendru/darwin/go/templater/internal/files"
	"github.com/otiai10/copy"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type Templater struct {
	evalCtx  *hcl.EvalContext
	manifest *Manifest
}

func New(evalCtx *hcl.EvalContext, manifest *Manifest) *Templater {
	return &Templater{
		evalCtx:  evalCtx,
		manifest: manifest,
	}
}

func (t *Templater) GetInputs() []*Input {
	return t.manifest.Inputs
}

func (t *Templater) Render(args map[string]interface{}, outDir string) error {
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return err
	}

	vals, err := t.getTemplateValues(args)
	if err != nil {
		return err
	}

	// Write everything to a temporary directory first, then copy it to the destination if there were no problems.
	tmpDir, err := ioutil.TempDir("", "templater")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	inDir := t.manifest.baseDir
	prefexLen := len(inDir)
	if err := filepath.Walk(inDir, func(inPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		outPath := path.Join(tmpDir, inPath[prefexLen:])
		// Allow file/direcory names to be templated.
		if strings.Contains(outPath, "{{") {
			if outPath, err = compileAndRunTemplate(string(outPath), vals); err != nil {
				return fmt.Errorf("evaluating template %s: %w", inPath, err)
			}
		}

		if info.IsDir() {
			os.Mkdir(outPath, info.Mode())
		} else {
			if info.Name() == "manifest.hcl" {
				return nil
			}

			if strings.HasSuffix(inPath, ".tpl") {
				outPath = outPath[:len(outPath)-4]
				return files.CopyWithTransform(inPath, outPath, info.Mode(), func(data []byte) ([]byte, error) {
					out, err := compileAndRunTemplate(string(data), vals)
					if err != nil {
						return nil, fmt.Errorf("evaluating template %s: %w", inPath, err)
					}

					return []byte(out), nil
				})
			} else {
				return files.Copy(inPath, outPath)
			}
		}

		return nil
	}); err != nil {
		log.Fatalln(err)
	}

	if err := copy.Copy(tmpDir, outDir); err != nil {
		return fmt.Errorf("copying template files to destination: %w", err)
	}

	return nil
}

func FromManifest(filename string) (*Templater, error) {
	evalCtx := DefaultEvalContext()
	manifest, err := ParseManifest(evalCtx, filename)
	if err != nil {
		return nil, err
	}

	return New(evalCtx, manifest), nil
}

func DefaultEvalContext() *hcl.EvalContext {
	return &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"env": envMap(),
		},
		Functions: map[string]function.Function{
			"upper":       stdlib.UpperFunc,
			"lower":       stdlib.LowerFunc,
			"min":         stdlib.MinFunc,
			"max":         stdlib.MaxFunc,
			"strlen":      stdlib.StrlenFunc,
			"substr":      stdlib.SubstrFunc,
			"kebab_case":  kebabCaseFunc,
			"slugify":     kebabCaseFunc, // alias for kebab_case
			"snake_case":  snakeCaseFunc,
			"camel_case":  camelCaseFunc,
			"pascal_case": pascalCaseFunc,
			"gitconfig":   gitconfigFunc,
		},
	}
}

func (t *Templater) getTemplateValues(args map[string]interface{}) (map[string]interface{}, error) {
	var err error
	values := map[string]interface{}{}

	vars := map[string]cty.Value{}
	for _, input := range t.manifest.Inputs {
		arg, ok := args[input.Name]
		var effectiveValue interface{}
		// Add the provided or default value to the variable map of the evaluation context so that it
		// may be referenced in an attribute.
		if !ok {
			if input.Required {
				return nil, fmt.Errorf("required input not supplied: %q", input.Name)
			}
			if input.DefaultValue != nil {
				var err error
				if vars[input.Name], err = input.CoerceTyped(input.DefaultValue); err != nil {
					return nil, err
				}
				effectiveValue = input.DefaultValue
			}
		} else {
			if vars[input.Name], err = input.CoerceTyped(arg); err != nil {
				return nil, err
			}
			effectiveValue = arg
		}

		// Also add the variable to the values available to the template.
		values[input.Name], err = input.Coerce(effectiveValue)
		if err != nil {
			panic("Coerce errored where CoerceTyped did not.")
		}
	}
	t.evalCtx.Variables["var"] = cty.MapVal(vars)

	// Compute attributes.
	var diags hcl.Diagnostics
	for _, a := range t.manifest.Attributes {
		diags = diags.Extend(a.evaluate(t.evalCtx))
		values[a.Name] = a.Val
	}
	if err := checkDiagnostics(diags); err != nil {
		return nil, err
	}

	return values, nil
}

func checkDiagnostics(diags hcl.Diagnostics) error {
	if !diags.HasErrors() {
		return nil
	}

	var sb strings.Builder
	isFirst := true
	for _, err := range diags.Errs() {
		if isFirst {
			isFirst = false
		} else {
			sb.WriteByte(' ')
		}
		sb.WriteString(err.Error())
		sb.WriteByte(';')
	}
	return fmt.Errorf("processing manifest: %s", sb.String())
}

func compileAndRunTemplate(content string, vars map[string]interface{}) (string, error) {
	tpl := template.New("template")
	parsed, err := tpl.Parse(content)
	if err != nil {
		return "", err
	}

	var out bytes.Buffer
	if err := parsed.Execute(&out, vars); err != nil {
		return "", err
	}

	return out.String(), nil
}
