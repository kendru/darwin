package evaluator

import (
	"fmt"
	"testing"

	"github.com/kendru/darwin/go/monkey/lexer"
	"github.com/kendru/darwin/go/monkey/object"
	"github.com/kendru/darwin/go/monkey/parser"
	"github.com/stretchr/testify/assert"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"-0", 0},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		_ = testIntegerObject(t, evaluated, tt.expected, fmt.Sprintf("expected eval: %q -> %d", tt.input, tt.expected))
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"
	value := testEval(input)
	fn, ok := value.(*object.Function)
	if !assert.True(t, ok, "expected function object") {
		return
	}

	assert.Len(t, fn.Parameters, 1)
	assert.Equal(t, "x", fn.Parameters[0].String())
	assert.Equal(t, "(x + 2)", fn.Body.String())
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x + x; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5);", 5},
		{"fn(a) { fn(b) { a + b } }(2)(3);", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected, fmt.Sprintf("expected function application eval: %q -> %d", tt.input, tt.expected))
	}
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64, errorMsg string) bool {
	res, ok := obj.(*object.Integer)
	return assert.True(t, ok, fmt.Sprintf("expected interger, but got %T: %s", obj, errorMsg)) &&
		assert.Equal(t, expected, res.Value, fmt.Sprintf("integer value not expected: %s", errorMsg))
}

func testNullObject(t *testing.T, obj object.Object) bool {
	return assert.Equal(t, NULL, obj)
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 == 2", false},
		{"1 != 2", true},
		{"1 != 1", false},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"(1 < 3) == true", true},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		_ = testBooleanObject(t, value, tt.expected, fmt.Sprintf("expected eval: %q -> %t", tt.input, tt.expected))
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool, errorMsg string) bool {
	res, ok := obj.(*object.Boolean)
	return assert.True(t, ok, "object not boolean") &&
		assert.Equal(t, expected, res.Value, fmt.Sprintf("boolean value not expected: %s", errorMsg))
}

func testStringObject(t *testing.T, obj object.Object, expected string, errorMsg string) bool {
	res, ok := obj.(*object.String)
	return assert.NotNil(t, obj) &&
		assert.Truef(t, ok, "expected string. got: %s", obj.Type()) &&
		assert.Equal(t, expected, res.Value, fmt.Sprintf("string value not expected: %s", errorMsg))
}

func testErrorObject(t *testing.T, obj object.Object, expectedMsg string) bool {
	errObj, ok := obj.(*object.Error)
	return assert.True(t, ok, "expected value to be error") &&
		assert.Equal(t, expectedMsg, errObj.Message)
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!0", false},
		{"!!true", true},
		{"!!false", false},
		{"!false", true},
		{"!!5", true},
		{"!!0", true},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		testBooleanObject(t, value, tt.expected, fmt.Sprintf("expected eval: %q -> %t", tt.input, tt.expected))
	}
}

func TestStringExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"Hello World";`, "Hello World"},
		{`"Marmalade" ++ " sandwich";`, "Marmalade sandwich"},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		testStringObject(t, value, tt.expected, fmt.Sprintf("expected eval: %q -> %s", tt.input, tt.expected))
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 }", 20},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, value, int64(expected), fmt.Sprintf("unexpected if/else result: %q -> %v", tt.input, tt.expected))
		default:
			testNullObject(t, value)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"if (true) { if (true) { return 10; } return 1; }", 10},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		testIntegerObject(t, value, tt.expected, fmt.Sprintf("expected return eval: %q -> %d", tt.input, tt.expected))
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("");`, 0},
		{`len("four");`, 4},
		{`len("hello world");`, 11},
		{`len([1,1,1]);`, 3},
		{`len(1);`, "argument to `len` not supported. got INTEGER"},
		{`len("one", "two");`, "wrong number of arguments. got=2, want=1"},
		{`head([1,2,3]);`, 1},
		{`head(tail([1,2,3]));`, 2},
		{`last([1,2,3]);`, 3},
		{`last(push([1,2,3], 4));`, 4},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, value, int64(expected), fmt.Sprintf("unexpected builtin result: %q -> %d", tt.input, expected))
		case string:
			testErrorObject(t, value, expected)
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"
	value := testEval(input)
	arr, ok := value.(*object.Array)
	if !assert.True(t, ok, "Value not an array") {
		return
	}
	if !assert.Len(t, arr.Elements, 3) {
		return
	}

	testIntegerObject(t, arr.Elements[0], 1, "")
	testIntegerObject(t, arr.Elements[1], 4, "")
	testIntegerObject(t, arr.Elements[2], 6, "")
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`[1,2,3][0]`, 1},
		{`[1,2,3][1]`, 2},
		{`[1,2,3][2]`, 3},
		{`[1,2,3][3]`, "array index out of bounds: 3"},
		{`[1,2,3][-1]`, "array index out of bounds: -1"},
		{`let i = 0; [1][i]`, 1},
		{`let a = [1,2,3]; a[2]`, 3},
		{`let a = [1,2,3]; a[0] + a[1] + a[2]`, 6},
		{`let a = [1,2,3]; let i = a[0]; a[i]`, 2},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, value, int64(expected), fmt.Sprintf("unexpected array index result: %q -> %d", tt.input, expected))
		case string:
			testErrorObject(t, value, expected)
		case nil:
			testNullObject(t, value)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `
  let two = "two";
  {
    "one": 10-9,
    two: 1+1,
    "thr" ++ "ee": 6/2,
    4: 4,
    true: 5,
    false: 6
  }
  `
	value := testEval(input)
	hash, ok := value.(*object.Hash)
	if !assert.True(t, ok, "Value not a hash") {
		return
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		(&object.Boolean{Value: true}).HashKey():   5,
		(&object.Boolean{Value: false}).HashKey():  6,
	}

	assert.Equal(t, len(expected), len(hash.Pairs), "hash length not correct")

	for expectedKey, expectedVal := range expected {
		pair, ok := hash.Pairs[expectedKey]
		if !assert.True(t, ok, "element not found in hash") {
			continue
		}
		testIntegerObject(t, pair.Value, expectedVal, "")
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"foo":5}["foo"]`, 5},
		{`{"foo":5}["bar"]`, nil},
		{`let key = "foo"; {"foo":5}[key]`, 5},
		{`{}["foo"]`, nil},
		{`{10:5}[10]`, 5},
		{`{true:5}[true]`, 5},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, value, int64(expected), fmt.Sprintf("unexpected array index result: %q -> %d", tt.input, expected))
		case nil:
			testNullObject(t, value)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
      if (10 > 1) {
        if (10 > 1) {
          return true + false;
        }
        return 1;
      }
      `,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"hello" - "world"`,
			"unknown operator: STRING - STRING",
		},
		{
			`"hello" + "world"`,
			"unknown operator: STRING + STRING",
		},
		{
			`{}[fn(x){x}]`,
			"key not hashable: FUNCTION",
		},
	}

	for _, tt := range tests {
		value := testEval(tt.input)
		if !testErrorObject(t, value, tt.expectedMessage) {
			continue
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tt := range tests {
		testIntegerObject(
			t,
			testEval(tt.input),
			tt.expected,
			fmt.Sprintf("expected let evaluation: %q -> %d", tt.input, tt.expected),
		)
	}
}

func testEval(input string) object.Object {
	l := lexer.New("<test>", input)
	p := parser.New(l)
	prog := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(prog, env)
}
