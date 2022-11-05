package inflect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInflect(t *testing.T) {
	tt := []struct {
		in     string
		camel  string
		pascal string
		snake  string
		kebab  string
	}{
		{
			in:     "Hello World",
			camel:  "helloWorld",
			pascal: "HelloWorld",
			snake:  "hello_world",
			kebab:  "hello-world",
		},
		{
			in:     "helloWorld",
			camel:  "helloWorld",
			pascal: "HelloWorld",
			snake:  "hello_world",
			kebab:  "hello-world",
		},
		{
			in:     "hello_world",
			camel:  "helloWorld",
			pascal: "HelloWorld",
			snake:  "hello_world",
			kebab:  "hello-world",
		},
		{
			in:     "Some_moreComplex__thing here",
			camel:  "someMoreComplexThingHere",
			pascal: "SomeMoreComplexThingHere",
			snake:  "some_more_complex_thing_here",
			kebab:  "some-more-complex-thing-here",
		},
	}

	i := Inflect{}
	for _, testCase := range tt {
		testCase := testCase
		t.Run(testCase.in, func(t *testing.T) {
			assert.Equal(t, testCase.camel, i.ToCamelCase(testCase.in), "camelCase")
			assert.Equal(t, testCase.pascal, i.ToPascalCase(testCase.in), "pascalCase")
			assert.Equal(t, testCase.snake, i.ToSnakeCase(testCase.in), "snake_case")
			assert.Equal(t, testCase.kebab, i.ToKebabCase(testCase.in), "kebab-case")
		})
	}
}
