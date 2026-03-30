package vm_test

import (
	"fmt"
	"monkey/ast"
	"monkey/compiler"
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"monkey/vm"
	"testing"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)

	return p.ParseProgram()
}

type vmTestCase struct {
	name     string
	input    string
	expected any
}

func TestIntegerArithmatic(t *testing.T) {
	tests := []vmTestCase{
		{name: "integer literal evaluates to one", input: "1", expected: 1},
		{name: "integer literal evaluates to two", input: "2", expected: 2},
		{name: "addition combines two integers", input: "1 + 2", expected: 3},
		{name: "subtraction produces a negative result", input: "1 - 2", expected: -1},
		{name: "multiplication combines two integers", input: "1 * 2", expected: 2},
		{name: "division returns the quotient", input: "4 / 2", expected: 2},
		{name: "mixed arithmetic respects operator precedence", input: "50 / 2 * 2 + 10 - 5", expected: 55},
		{name: "parentheses override arithmetic precedence", input: "5 * (2 + 10)", expected: 60},
		{name: "chained additions and subtraction evaluate left to right", input: "5 + 5 + 5 + 5 - 10", expected: 10},
		{name: "repeated multiplication accumulates correctly", input: "2 * 2 * 2 * 2 * 2", expected: 32},
		{name: "multiplication is applied before trailing addition", input: "5 * 2 + 10", expected: 20},
		{name: "multiplication has higher precedence than addition", input: "5 + 2 * 10", expected: 25},
		{name: "grouped sum is multiplied before returning", input: "5 * (2 + 10)", expected: 60},
		{name: "unary minus negates a single digit integer", input: "-5", expected: -5},
		{name: "unary minus negates a multi digit integer", input: "-10", expected: -10},
		{name: "negative operands participate in addition", input: "-50 + 100 + -50", expected: 0},
		{name: "nested arithmetic handles precedence and unary minus together", input: "(5+10*2+15/3)*2 + -10", expected: 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{name: "boolean literal true evaluates truthy", input: "true", expected: true},
		{name: "boolean literal false evaluates falsy", input: "false", expected: false},
		{name: "less than comparison returns true", input: "1 < 2", expected: true},
		{name: "greater than comparison returns false", input: "1 > 2", expected: false},
		{name: "less than comparison returns false for equal values", input: "1 < 1", expected: false},
		{name: "greater than comparison returns false for equal values", input: "1 > 1", expected: false},
		{name: "integer equality returns true for identical values", input: "1 == 1", expected: true},
		{name: "integer inequality returns false for identical values", input: "1 != 1", expected: false},
		{name: "integer equality returns false for different values", input: "1 == 2", expected: false},
		{name: "integer inequality returns true for different values", input: "1 != 2", expected: true},
		{name: "boolean equality returns true for true values", input: "true == true", expected: true},
		{name: "boolean equality returns true for false values", input: "false == false", expected: true},
		{name: "boolean equality returns false for mismatched values", input: "true == false", expected: false},
		{name: "boolean inequality returns true for true and false", input: "true != false", expected: true},
		{name: "boolean inequality returns true for false and true", input: "false != true", expected: true},
		{name: "comparison result can be compared to true", input: "(1 < 2) == true", expected: true},
		{name: "comparison result can be compared to false", input: "(1 < 2) == false", expected: false},
		{name: "false comparison result is not equal to true", input: "(1 > 2) == true", expected: false},
		{name: "false comparison result is equal to false", input: "(1 > 2) == false", expected: true},
		{name: "bang operator inverts false", input: "!false", expected: true},
		{name: "double bang preserves false", input: "!!false", expected: false},
		{name: "bang operator inverts true", input: "!true", expected: false},
		{name: "double bang preserves true", input: "!!true", expected: true},
		{name: "bang operator treats integers as truthy", input: "!5", expected: false},
		{name: "double bang converts integers to truthy booleans", input: "!!5", expected: true},
		{name: "bang operator treats null if results as falsy", input: "!(if (false) { 5; })", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{name: "if executes consequence for true condition", input: "if (true) { 10 }", expected: 10},
		{name: "if else returns consequence when condition is true", input: "if (true) { 10 } else { 20 }", expected: 10},
		{name: "if else returns alternative when condition is false", input: "if (false) { 10 } else { 20 } ", expected: 20},
		{name: "if treats non boolean integers as truthy", input: "if (1) { 10 }", expected: 10},
		{name: "if executes consequence for true comparison", input: "if (1 < 2) { 10 }", expected: 10},
		{name: "if else returns consequence for true comparison", input: "if (1 < 2) { 10 } else { 20 }", expected: 10},
		{name: "if else returns alternative for false comparison", input: "if (1 > 2) { 10 } else { 20 }", expected: 20},
		{name: "if without alternative returns null for false comparison", input: "if (1 > 2) { 10 }", expected: vm.Null},
		{name: "if without alternative returns null for false literal", input: "if (false) { 10 }", expected: vm.Null},
		{name: "if treats nested null condition as falsy", input: "if ((if (false) { 10 })) { 10 } else { 20 }", expected: 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []vmTestCase{
		{name: "global let binding can be read back", input: "let one = 1; one", expected: 1},
		{name: "multiple global let bindings can be combined", input: "let one = 1; let two = 2; one + two", expected: 3},
		{name: "global let bindings can reference earlier bindings", input: "let one = 1; let two = one + one; one + two", expected: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{
			name:     "string literal evaluates to string object",
			input:    `"monkey"`,
			expected: "monkey",
		},
		{
			name:     "string concatenation joins two strings",
			input:    `"mon"+"key"`,
			expected: "monkey",
		},
		{
			name:     "string concatenation supports multiple operands",
			input:    `"mon"+"key"+"banana"`,
			expected: "monkeybanana",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestArrayLiterals(t *testing.T) {
	tests := []vmTestCase{
		{name: "array literal supports empty arrays", input: "[]", expected: []int{}},
		{name: "array literal preserves integer elements", input: "[1, 2, 3]", expected: []int{1, 2, 3}},
		{name: "array literal evaluates element expressions", input: "[1 + 2, 3 * 4, 5 + 6]", expected: []int{3, 12, 11}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestHashLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			name:     "hash literal supports empty hashes",
			input:    "{}",
			expected: map[object.HashKey]int64{},
		},
		{
			name:  "hash literal stores integer keyed pairs",
			input: "{1: 2, 2: 3}",
			expected: map[object.HashKey]int64{
				(&object.Integer{Value: 1}).HashKey(): 2,
				(&object.Integer{Value: 2}).HashKey(): 3,
			},
		},
		{
			name:  "hash literal evaluates computed keys and values",
			input: "{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			expected: map[object.HashKey]int64{
				(&object.Integer{Value: 2}).HashKey(): 4,
				(&object.Integer{Value: 6}).HashKey(): 16,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{name: "array indexing reads an existing element", input: "[1, 2, 3][1]", expected: 2},
		{name: "array indexing evaluates index expressions", input: "[1, 2, 3][0 + 2]", expected: 3},
		{name: "nested array indexing can be chained", input: "[[1, 1, 1]][0][0]", expected: 1},
		{name: "array indexing returns null for empty arrays", input: "[][0]", expected: vm.Null},
		{name: "array indexing returns null when out of bounds", input: "[1, 2, 3][99]", expected: vm.Null},
		{name: "array indexing returns null for negative indexes", input: "[1][-1]", expected: vm.Null},
		{name: "hash indexing reads the first existing key", input: "{1: 1, 2: 2}[1]", expected: 1},
		{name: "hash indexing reads the second existing key", input: "{1: 1, 2: 2}[2]", expected: 2},
		{name: "hash indexing returns null for missing keys", input: "{1: 1}[0]", expected: vm.Null},
		{name: "hash indexing returns null for empty hashes", input: "{}[0]", expected: vm.Null},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			name:     "function call returns the last evaluated expression",
			input:    `let f = fn() { 5 + 10; }; f();`,
			expected: 15,
		},
		{
			name:     "function call respects explicit return statements",
			input:    `let f = fn() { return 15; }; f();`,
			expected: 15,
		},
		{
			name:     "function call stops executing after return",
			input:    `let f = fn() { return 15; return 10; }; f();`,
			expected: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestFunctionsWithoutReturns(t *testing.T) {
	tests := []vmTestCase{
		{
			name:     "function call returns null for empty bodies",
			input:    `let f = fn() { }; f();`,
			expected: vm.Null,
		},
		{
			name: "function calls propagate null when no return value exists",
			input: `
let noReturn = fn() { }; 
let noReturnAgain = fn () { noReturn(); }; 
noReturn(); 
noReturnAgain();
			`,
			expected: vm.Null,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			name: "first class functions can return callable functions",
			input: `
let r = fn() {1;};
let rReturner = fn() { r; };
rReturner()();
`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			name: "functions can read their own local bindings",
			input: `
		let one = fn() { let one = 1; one };
		one();
		`,
			expected: 1,
		},
		{
			name: "functions can combine multiple local bindings",
			input: `
		let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
		oneAndTwo();
		`,
			expected: 3,
		},
		{
			name: "different functions keep local bindings isolated",
			input: `
		let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
		let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
		oneAndTwo() + threeAndFour();
		`,
			expected: 10,
		},
		{
			name: "same local names do not leak across functions",
			input: `
		let firstFoobar = fn() { let foobar = 50; foobar; };
		let secondFoobar = fn() { let foobar = 100; foobar; };
		firstFoobar() + secondFoobar();
		`,
			expected: 150,
		},
		{
			name: "functions can read global bindings alongside locals",
			input: `
		let globalSeed = 50;
		let minusOne = fn() {
			let num = 1;
			globalSeed - num;
		}
		let minusTwo = fn() {
			let num = 2;
			globalSeed - num;
		}
		minusOne() + minusTwo();
		`,
			expected: 97,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestCallingFunctionsWithArgsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			name:     "functions can use args in their local scope",
			input:    `let identity = fn(a) { a; }; identity(4);`,
			expected: 4,
		},
		{
			name:     "functions can take multiple arguments and perform operations with them",
			input:    `let sum = fn(a,b) { a+b; }; sum(1,2);`,
			expected: 3,
		},
		{
			name: "mix manually created local bindings with args and call multiple times",
			input: `
		let sum = fn(a, b) {
			let c = a + b;
			c;
		};
		sum(1, 2) + sum(3, 4);`,
			expected: 10,
		},
		{
			name: "mix manually created local bindings with args and call inside another function",
			input: `
		let sum = fn(a, b) {
			let c = a + b;
			c;
		};
		let outer = fn() {
			sum(1, 2) + sum(3, 4);
		};
		outer();
		`,
			expected: 10,
		},
		{
			name: "mix global bindings, local bindings, functions and outer functions",
			input: `
		let globalNum = 10;

		let sum = fn(a, b) {
			let c = a + b;
			c + globalNum;
		};

		let outer = fn() {
			sum(1, 2) + sum(3, 4) + globalNum;
		};

		outer() + globalNum;
		`,
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			name:     "calling noArgs function with one args",
			input:    `fn() { 1; }(1);`,
			expected: `wrong number of arguments: want=0, got=1`,
		},
		{
			name:     "calling args function with no args",
			input:    `fn(a) { a; }();`,
			expected: `wrong number of arguments: want=1, got=0`,
		},
		{
			name:     "calling multi args function with one arg",
			input:    `fn(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: want=2, got=1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program := parse(tt.input)

			comp := compiler.New()
			err := comp.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %s", err)
			}

			virtualMachine := vm.New(comp.Bytecode())
			err = virtualMachine.Run()
			if err == nil {
				t.Fatalf("expected VM error but resulted in none.")
			}

			if err.Error() != tt.expected {
				t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err)
			}
		})

	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			name:     "valid len() on empty string",
			input:    `len("")`,
			expected: 0,
		},
		{
			name:     "valid len() on single word string",
			input:    `len("four")`,
			expected: 4,
		},
		{
			name:     "valid len() on space separated string",
			input:    `len("Hello World!")`,
			expected: 12,
		},
		{
			name:     "invalid len() on number",
			input:    `len(1)`,
			expected: &object.Error{Message: "argument to `len` not supported, got INTEGER"},
		},
		{
			name:     "invalid len() on string: wrong number of args",
			input:    `len("one", "two")`,
			expected: &object.Error{Message: "wrong number of arguments. got=2, want=1"},
		},
		{
			name:     "push(array, 1) does not persist the update",
			input:    `let a = [1]; push(a, 2); a`,
			expected: []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVmTest(t, tt)
		})
	}
}

func TestBuiltinFunctionsREPL(t *testing.T) {
	tests := []struct {
		inputs   []string
		expected any
	}{
		{
			inputs: []string{
				"let arr = [1, 2, 3];",
				"push(arr, 4);",
				"arr",
			},
			expected: []int{1, 2, 3, 4},
		},
	}

	for _, tt := range tests {
		constants := []object.Object{}
		globals := make([]object.Object, vm.GlobalSize)
		symTable := compiler.NewSymbolTable()
		for i, v := range object.Builtins {
			symTable.DefineBuiltin(i, v.Name)
		}

		var lastPopped object.Object
		for _, input := range tt.inputs {
			program := parse(input)
			comp := compiler.NewWithState(symTable, constants)
			err := comp.Compile(program)
			if err != nil {
				t.Fatalf("compiler error: %s", err)
			}

			machine := vm.NewWithGlobalStore(comp.Bytecode(), globals)
			err = machine.Run()
			if err != nil {
				t.Fatalf("vm error: %s", err)
			}
			lastPopped = machine.LastPoppedStackElem()
			constants = comp.Bytecode().Constants
		}

		testExpectedObject(t, tt.expected, lastPopped)
	}
}

func runVmTest(t *testing.T, testCase vmTestCase) {
	t.Helper()
	program := parse(testCase.input)

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	virtualMachine := vm.New(comp.Bytecode())
	err = virtualMachine.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	stackElem := virtualMachine.LastPoppedStackElem()

	testExpectedObject(t, testCase.expected, stackElem)
}

func testExpectedObject(t *testing.T, expected any, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}

	case bool:
		err := testBooleanObject(expected, actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}

	case string:
		err := testStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed:%s", err)
		}

	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object not Array: %T (%+v)", actual, actual)
			return
		}

		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d",
				len(expected), len(array.Elements))
			return
		}

		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case map[object.HashKey]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d",
				len(expected), len(hash.Pairs))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := hash.Pairs[expectedKey]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}

			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case *object.Null:
		if actual != vm.Null {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}

	case *object.Error:
		errorObject, ok := actual.(*object.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", actual, actual)
			return
		}
		if errorObject.Message != expected.Message {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.Message, errorObject.Message)
		}
	}
}

func testStringObject(expected string, actual object.Object) error {
	result, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not a String. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. want=%q, got=%q", expected, result.Value)
	}

	return nil
}

func testBooleanObject(expected bool, actual object.Object) error {
	res, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}

	if res.Value != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", res.Value, expected)
	}
	return nil
}

func testIntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not integer type. got=%T (%+v)", actual, actual)
	}

	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
	}

	return nil
}
