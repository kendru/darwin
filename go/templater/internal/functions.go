package templater

import (
	"log"
	"os/exec"
	"strings"

	"github.com/kendru/darwin/go/templater/internal/inflect"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// This file contains the implementation for the HCL functions that are added to the
// execution context.

var kebabCaseFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: kebabCaseImpl,
})

func kebabCaseImpl(args []cty.Value, retType cty.Type) (cty.Value, error) {
	str := args[0].AsString()
	str = inflect.Default.ToKebabCase(str)

	return cty.StringVal(string(str)), nil
}

var snakeCaseFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: snakeCaseImpl,
})

func snakeCaseImpl(args []cty.Value, retType cty.Type) (cty.Value, error) {
	str := args[0].AsString()
	str = inflect.Default.ToSnakeCase(str)

	return cty.StringVal(string(str)), nil
}

var camelCaseFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: camelCaseImpl,
})

func camelCaseImpl(args []cty.Value, retType cty.Type) (cty.Value, error) {
	str := args[0].AsString()
	str = inflect.Default.ToCamelCase(str)

	return cty.StringVal(string(str)), nil
}

var pascalCaseFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: pascalCaseImpl,
})

func pascalCaseImpl(args []cty.Value, retType cty.Type) (cty.Value, error) {
	str := args[0].AsString()
	str = inflect.Default.ToPascalCase(str)

	return cty.StringVal(string(str)), nil
}

var gitconfigFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name:             "key",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: gitconfigImpl,
})

func gitconfigImpl(args []cty.Value, retType cty.Type) (cty.Value, error) {
	key := args[0].AsString()
	out, err := exec.Command("git", "config", key).Output()
	if err != nil {
		log.Fatal(err)
	}
	str := string(out)
	str = strings.Trim(str, " \t\n\r")

	return cty.StringVal(str), nil
}
