package parser

import (
	"fmt"
	"testing"

	"github.com/kendru/darwin/go/monkey/ast"
	"github.com/kendru/darwin/go/monkey/lexer"
	"github.com/stretchr/testify/assert"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		prog := parseAndCheckErrors(t, tt.input)
		if !assert.Len(t, prog.Statements, 1) {
			return
		}

		stmt := prog.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier) {
			return
		}

		val := stmt.(*ast.LetStatement).Value
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func testLetStatement(t *testing.T, s ast.Statement, name string) bool {
	if !assert.Equal(t, "let", s.TokenLiteral()) {
		return false
	}

	letStmt, ok := s.(*ast.LetStatement)
	if !assert.True(t, ok) {
		return false
	}

	if !assert.Equal(t, name, letStmt.Name.Value) {
		return false
	}

	if !assert.Equal(t, name, letStmt.Name.TokenLiteral()) {
		return false
	}

	return true
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return true;", true},
		{"return y;", "y"},
	}

	for _, tt := range tests {
		prog := parseAndCheckErrors(t, tt.input)
		if !assert.Len(t, prog.Statements, 1) {
			return
		}

		stmt := prog.Statements[0]
		if !testReturnStatement(t, stmt) {
			return
		}

		val := stmt.(*ast.ReturnStatement).ReturnValue
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func testReturnStatement(t *testing.T, stmt ast.Statement) bool {
	stmt, ok := stmt.(*ast.ReturnStatement)
	return assert.True(t, ok) &&
		assert.Equal(t, "return", stmt.TokenLiteral())
}

func TestIdentifierExpression(t *testing.T) {
	input := "tacos;"

	prog := parseAndCheckErrors(t, input)
	if !assert.Equal(t, 1, len(prog.Statements)) {
		return
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)
	testIdentifier(t, stmt.Expression, "tacos")
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "7;"

	prog := parseAndCheckErrors(t, input)
	if !assert.Equal(t, 1, len(prog.Statements)) {
		return
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	testIntegerLiteral(t, stmt.Expression, int64(7))
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue bool
	}{
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		prog := parseAndCheckErrors(t, tt.input)
		if !assert.Equal(t, 1, len(prog.Statements)) {
			return
		}

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok)

		testBooleanLiteral(t, stmt.Expression, tt.expectedValue)
	}
}

func TestStringLiteralExpression(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue string
	}{
		{`"nachos";`, "nachos"},
	}

	for _, tt := range tests {
		prog := parseAndCheckErrors(t, tt.input)
		if !assert.Equal(t, 1, len(prog.Statements)) {
			return
		}

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok)

		testStringLiteral(t, stmt.Expression, tt.expectedValue)
	}
}

func TestPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
	}

	for _, tt := range prefixTests {
		prog := parseAndCheckErrors(t, tt.input)
		if !assert.Equal(t, 1, len(prog.Statements)) {
			return
		}

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok)

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		assert.True(t, ok)
		assert.Equal(t, tt.operator, exp.Operator)
		if !testIntegerLiteral(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

func TestInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		operator   string
		leftValue  int64
		rightValue int64
	}{
		{"5 + 5;", "+", 5, 5},
		{"5 - 5;", "-", 5, 5},
		{"5 * 5;", "*", 5, 5},
		{"5 / 5;", "/", 5, 5},
		{"5 > 5;", ">", 5, 5},
		{"5 < 5;", "<", 5, 5},
		{"5 == 5;", "==", 5, 5},
		{"5 != 5;", "!=", 5, 5},
	}

	for _, tt := range infixTests {
		prog := parseAndCheckErrors(t, tt.input)

		if !assert.Equal(t, 1, len(prog.Statements)) {
			return
		}

		stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
		assert.True(t, ok)

		if !testInfixExpression(t, stmt.Expression, tt.operator, tt.leftValue, tt.rightValue) {
			return
		}
	}
}

func TestIfExpression(t *testing.T) {
	input := "if (x < y) { x }"
	prog := parseAndCheckErrors(t, input)
	if !assert.Len(t, prog.Statements, 1) {
		return
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	exp, ok := stmt.Expression.(*ast.IfExpression)
	assert.True(t, ok)

	if !testInfixExpression(t, exp.Condition, "<", "x", "y") {
		return
	}

	if !assert.Len(t, exp.Consequence.Statements, 1) {
		return
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	assert.Nil(t, exp.Alternative)
}

func TestIfElseExpression(t *testing.T) {
	input := "if (x < y) { x } else { y }"
	prog := parseAndCheckErrors(t, input)
	if !assert.Len(t, prog.Statements, 1) {
		return
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	exp, ok := stmt.Expression.(*ast.IfExpression)
	assert.True(t, ok)

	if !testInfixExpression(t, exp.Condition, "<", "x", "y") {
		return
	}

	if !assert.Len(t, exp.Consequence.Statements, 1) {
		return
	}

	consequence, ok := exp.Consequence.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	if !testIdentifier(t, consequence.Expression, "x") {
		return
	}

	alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
	assert.True(t, ok)

	if !testIdentifier(t, alternative.Expression, "y") {
		return
	}
}

func TestFunctionLiteralExpression(t *testing.T) {
	input := "fn(x, y) { x + y; }"
	prog := parseAndCheckErrors(t, input)
	if !assert.Len(t, prog.Statements, 1) {
		return
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	if !assert.True(t, ok) {
		return
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !assert.True(t, ok) {
		return
	}

	if !assert.Len(t, function.Parameters, 2) {
		return
	}

	if !(testLiteralExpression(t, function.Parameters[0], "x") &&
		testLiteralExpression(t, function.Parameters[1], "y")) {
		return
	}

	if !assert.Len(t, function.Body.Statements, 1) {
		return
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !assert.True(t, ok) {
		return
	}

	testInfixExpression(t, bodyStmt.Expression, "+", "x", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{"fn(){};", []string{}},
		{"fn(x){};", []string{"x"}},
		{"fn(x,y, z){};", []string{"x", "y", "z"}},
	}

	for _, tt := range tests {
		prog := parseAndCheckErrors(t, tt.input)
		function := prog.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral)
		assert.Len(t, function.Parameters, len(tt.expectedParams))
		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestCallExpression(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"
	prog := parseAndCheckErrors(t, input)
	if !assert.Len(t, prog.Statements, 1) {
		return
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	if !assert.True(t, ok) {
		return
	}

	call, ok := stmt.Expression.(*ast.CallExpression)
	if !assert.True(t, ok) {
		return
	}

	if !testIdentifier(t, call.Function, "add") {
		return
	}

	if !assert.Len(t, call.Arguments, 3) {
		return
	}

	testLiteralExpression(t, call.Arguments[0], 1)
	testInfixExpression(t, call.Arguments[1], "*", 2, 3)
	testInfixExpression(t, call.Arguments[2], "+", 4, 5)
}

func TestArrayLiteralExpression(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"
	prog := parseAndCheckErrors(t, input)
	if !assert.Len(t, prog.Statements, 1) {
		return
	}

	stmt, ok := prog.Statements[0].(*ast.ExpressionStatement)
	if !assert.True(t, ok) {
		return
	}

	arr, ok := stmt.Expression.(*ast.ArrayLiteral)
	if !assert.True(t, ok) {
		return
	}

	if !assert.Len(t, arr.Elements, 3) {
		return
	}

	testLiteralExpression(t, arr.Elements[0], 1)
	testInfixExpression(t, arr.Elements[1], "*", 2, 2)
	testInfixExpression(t, arr.Elements[2], "+", 3, 3)
}

func TestIndexExpressions(t *testing.T) {
	input := "myArray[1 + 1]"
	prog := parseAndCheckErrors(t, input)
	if !assert.Len(t, prog.Statements, 1) {
		return
	}

	stmt, _ := prog.Statements[0].(*ast.ExpressionStatement)
	indexExp, ok := stmt.Expression.(*ast.IndexExpression)
	if !assert.True(t, ok) {
		return
	}

	if !testIdentifier(t, indexExp.Left, "myArray") {
		return
	}

	if !testInfixExpression(t, indexExp.Index, "+", 1, 1) {
		return
	}
}

func TestEmptyHashLiterals(t *testing.T) {
	input := `{}`
	prog := parseAndCheckErrors(t, input)
	stmt, _ := prog.Statements[0].(*ast.ExpressionStatement)
	exp, ok := stmt.Expression.(*ast.HashLiteral)
	if !assert.True(t, ok, "expected hash literal") {
		return
	}

	assert.Len(t, exp.Pairs, 0)
}

func TestHashLiteralsStringKeys(t *testing.T) {
	input := `{"one": 1, "two": 2, "three": 3}`
	prog := parseAndCheckErrors(t, input)
	stmt, _ := prog.Statements[0].(*ast.ExpressionStatement)
	exp, ok := stmt.Expression.(*ast.HashLiteral)
	if !assert.True(t, ok, "expected hash literal") {
		return
	}

	assert.Len(t, exp.Pairs, 3)

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
	}

	for key, value := range exp.Pairs {
		literal, ok := key.(*ast.StringLiteral)
		if !assert.True(t, ok, "expected string key") {
			return
		}
		expectedValue := expected[literal.String()]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestHashLiteralsIntegerKeys(t *testing.T) {
	input := `{10: 1, 20: 2, 30: 3}`
	prog := parseAndCheckErrors(t, input)
	stmt, _ := prog.Statements[0].(*ast.ExpressionStatement)
	exp, ok := stmt.Expression.(*ast.HashLiteral)
	if !assert.True(t, ok, "expected hash literal") {
		return
	}

	assert.Len(t, exp.Pairs, 3)

	expected := map[int64]int64{
		10: 1,
		20: 2,
		30: 3,
	}

	for key, value := range exp.Pairs {
		literal, ok := key.(*ast.IntegerLiteral)
		if !assert.True(t, ok, "expected integer key") {
			return
		}
		expectedValue := expected[literal.Value]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func TestHashLiteralsBooleanKeys(t *testing.T) {
	input := `{true: 1, false: 2}`
	prog := parseAndCheckErrors(t, input)
	stmt, _ := prog.Statements[0].(*ast.ExpressionStatement)
	exp, ok := stmt.Expression.(*ast.HashLiteral)
	if !assert.True(t, ok, "expected hash literal") {
		return
	}

	assert.Len(t, exp.Pairs, 2)

	expected := map[bool]int64{
		true:  1,
		false: 2,
	}

	for key, value := range exp.Pairs {
		literal, ok := key.(*ast.Boolean)
		if !assert.True(t, ok, "expected boolean key") {
			return
		}
		expectedValue := expected[literal.Value]
		testIntegerLiteral(t, value, expectedValue)
	}
}

func testInfixExpression(
	t *testing.T,
	exp ast.Expression,
	operator string,
	left, right interface{},
) bool {
	opExp, ok := exp.(*ast.InfixExpression)

	return assert.True(t, ok) &&
		testLiteralExpression(t, opExp.Left, left) &&
		assert.Equal(t, opExp.Operator, operator) &&
		testLiteralExpression(t, opExp.Right, right)
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, int64(v))
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}

	t.Errorf("unhandled expression type. got %T", exp)
	return false
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) bool {
	integ, ok := exp.(*ast.IntegerLiteral)

	return assert.True(t, ok) &&
		assert.Equal(t, value, integ.Value) &&
		assert.Equal(t, fmt.Sprintf("%d", value), integ.TokenLiteral())
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	b, ok := exp.(*ast.Boolean)

	return assert.True(t, ok) &&
		assert.Equal(t, value, b.Value) &&
		assert.Equal(t, fmt.Sprintf("%t", value), b.TokenLiteral())
}

func testStringLiteral(t *testing.T, exp ast.Expression, value string) bool {
	s, ok := exp.(*ast.StringLiteral)

	return assert.True(t, ok) &&
		assert.Equal(t, value, s.Value) &&
		assert.Equal(t, value, s.TokenLiteral())
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)

	return assert.True(t, ok) &&
		assert.Equal(t, value, ident.Value) &&
		assert.Equal(t, value, ident.TokenLiteral())
}

func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d)",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)))",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g))",
		},
		{
			"a * [1, 2, 3, 4][b * c] * d",
			"((a * ([1, 2, 3, 4][(b * c)])) * d)",
		},
		{
			"add(a * b[2], b[1], 2 * [1, 2][1])",
			"add((a * (b[2])), (b[1]), (2 * ([1, 2][1])))",
		},
	}

	for _, tt := range tests {
		l := lexer.New("<test>", tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		assert.Equal(t, tt.expected, actual)
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("encountered %d errors parsing", len(errors))
	for i, msg := range errors {
		t.Errorf("[%d] error: %q", i, msg)
	}
	t.FailNow()
}

func parseAndCheckErrors(t *testing.T, input string) *ast.Program {
	l := lexer.New("<test>", input)
	p := New(l)
	prog := p.ParseProgram()
	checkParserErrors(t, p)

	return prog
}
