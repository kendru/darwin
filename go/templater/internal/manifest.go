package templater

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

type Manifest struct {
	Inputs     []*Input     `hcl:"input,block"`
	Attributes []*Attribute `hcl:"attribute,block"`

	baseDir string
}

type Input struct {
	Name        string  `hcl:"name,label"`
	Description *string `hcl:"description"`
	Type        string  `hcl:"type"`
	Required    bool    `hcl:"required,optional"`

	DynamicConfig hcl.Attributes `hcl:",remain"`
	DefaultValue  interface{}
}

func ParseManifest(evalCtx *hcl.EvalContext, filename string) (*Manifest, error) {
	var manifest Manifest

	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(filename)
	if err := checkDiagnostics(diags); err != nil {
		return nil, err
	}

	diags = gohcl.DecodeBody(f.Body, evalCtx, &manifest)
	if err := checkDiagnostics(diags); err != nil {
		return nil, err
	}

	for _, v := range manifest.Inputs {
		diags = diags.Extend(v.processDynamicConfig(evalCtx))
	}
	if err := checkDiagnostics(diags); err != nil {
		return nil, err
	}

	// Store the directory that contained the manifest so that it can be read when rendering a template.
	manifest.baseDir = filepath.Dir(filename)

	return &manifest, nil
}

func (v *Input) processDynamicConfig(evalCtx *hcl.EvalContext) (diags hcl.Diagnostics) {
	for _, attr := range v.DynamicConfig {
		switch attr.Name {
		case "default":
			val, moreDiags := attr.Expr.Value(evalCtx)
			if !moreDiags.HasErrors() {
				diags = diags.Extend(v.setDefault(attr, val))
			}
			diags = diags.Extend(moreDiags)
		}
	}
	v.DynamicConfig = nil

	return
}

// Coerces attempts to coerce `val` to a type that satisfies the input.
func (i *Input) Coerce(val interface{}) (interface{}, error) {
	switch i.Type {
	case "string":
		return coerceString(val)
	case "int":
		return coerceInt(val)
	default:
		panic(fmt.Sprintf("invalid input type: %q", i.Type))
	}
}

func (i *Input) CoerceTyped(val interface{}) (cty.Value, error) {
	switch i.Type {
	case "string":
		str, err := coerceString(val)
		if err != nil {
			return cty.NilVal, err
		}
		return cty.StringVal(str), nil

	case "int":
		i, err := coerceInt(val)
		if err != nil {
			return cty.NilVal, err
		}
		return cty.NumberIntVal(i), nil

	default:
		panic(fmt.Sprintf("invalid input type: %q", i.Type))
	}
}

func (v *Input) setDefault(attr *hcl.Attribute, val cty.Value) (diags hcl.Diagnostics) {
	switch v.Type {
	case "string":
		if val.Type() == cty.String {
			v.DefaultValue = val.AsString()
		} else {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid type for default value",
				Detail:   fmt.Sprintf("Expected string but got %s", val.Type().FriendlyName()),
				Subject:  &attr.Range,
			})
		}
	}

	return
}

type Attribute struct {
	Name        string         `hcl:"name,label"`
	ValDeferred hcl.Attributes `hcl:",remain"`
	Val         interface{}
}

func (a *Attribute) evaluate(evalCtx *hcl.EvalContext) hcl.Diagnostics {
	for _, attr := range a.ValDeferred {
		if attr.Name == "val" {
			val, diags := attr.Expr.Value(evalCtx)
			switch val.Type() {
			case cty.Bool:
				a.Val = val.True()
			case cty.String:
				a.Val = val.AsString()
			case cty.Number:
				bf := val.AsBigFloat()
				// TODO: Handle int or float.
				a.Val, _ = bf.Int64()
			default:
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "unsupported attribute type",
					Detail:   fmt.Sprintf("Attribute %s is of unsupported type: %s", a.Name, val.Type().FriendlyName()),
				})
			}
			a.ValDeferred = nil
			return diags
		}
	}
	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "attribute val not specified",
			Detail:   fmt.Sprintf("Attribute %s missing \"val\" expression", a.Name),
		},
	}
}

func envMap() cty.Value {
	envEntries := os.Environ()
	envVars := make(map[string]cty.Value, len(envEntries))
	for _, e := range envEntries {
		pair := strings.SplitN(e, "=", 2)
		key := pair[0]
		val := pair[1]
		envVars[key] = cty.StringVal(val)
	}
	return cty.MapVal(envVars)
}
