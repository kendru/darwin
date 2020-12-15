package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/kendru/darwin/go/monkey/object"

	"github.com/kendru/darwin/go/monkey/evaluator"
	"github.com/kendru/darwin/go/monkey/lexer"
	"github.com/kendru/darwin/go/monkey/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New("<repl>", line)
		p := parser.New(l)

		prog := p.ParseProgram()
		if len(p.Errors()) > 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		value := evaluator.Eval(prog, env)
		if value != nil {
			io.WriteString(out, value.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "ERRORS:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t")
		io.WriteString(out, msg)
		io.WriteString(out, "\n")
	}
}
