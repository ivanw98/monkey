// Package repl tokenizes Monkey source code and prints the tokens.
package repl

import (
	"bufio"
	"fmt"
	"io"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/parser"
	"monkey/vm"
)

const PROMPT = ">>"

// Start reads from the input source.
func Start(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	//	env := object.NewEnvironment()

	for {
		_, err := fmt.Fprintf(out, PROMPT)
		if err != nil {
			return err
		}
		scanned := scanner.Scan()
		if !scanned {
			return nil
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		comp := compiler.New()
		err = comp.Compile(program)
		if err != nil {
			_, err = fmt.Fprintf(out, "Whoops! Compilation failed:\n %s\n", err)
			if err != nil {
				return err
			}
			continue
		}

		machine := vm.New(comp.Bytecode())
		err = machine.Run()
		if err != nil {
			_, err = fmt.Fprintf(out, "Whoops! Executing bytecode failed:\n %s\n", err)
			if err != nil {
				return err
			}
			continue
		}

		stackTop := machine.StackTop()
		_, err = io.WriteString(out, stackTop.Inspect())
		if err != nil {
			return err
		}

		_, err = io.WriteString(out, "\n")
		if err != nil {
			return err
		}

		//evaluated := evaluator.Eval(program, env)
		//if evaluated != nil {
		//	_, err = io.WriteString(out, evaluated.Inspect())
		//	if err != nil {
		//		return err
		//	}
		//	_, err = io.WriteString(out, "\n")
		//	if err != nil {
		//		return err
		//	}
		//}
	}
}

const MONKEY_FACE = `            __,__
   .--.  .-"     "-.  .--.
  / .. \/  .-. .-.  \/ .. \
 | |  '|  /   Y   \  |'  | |
 | \   \  \ 0 | 0 /  /   / |
  \ '- ,\.-"""""""-./, -' /
   ''-' /_   ^ ^   _\ '-''
       |  \._   _./  |
       \   \ '~' /   /
        '._ '-=-' _.'
           '-----'
`

func printParserErrors(out io.Writer, errors []string) {
	_, err := io.WriteString(out, MONKEY_FACE)
	if err != nil {
		return
	}
	_, err = io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	if err != nil {
		return
	}
	_, err = io.WriteString(out, " parser errors:\n")
	if err != nil {
		return
	}
	for _, msg := range errors {
		_, err = io.WriteString(out, "\t"+msg+"\n")
		if err != nil {
			return
		}
	}
}
