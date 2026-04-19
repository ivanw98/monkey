package compiler

import (
	"testing"
)

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GlobalScope, Index: 0},
		"b": {Name: "b", Scope: GlobalScope, Index: 1},
		"c": {Name: "c", Scope: LocalScope, Index: 0},
		"d": {Name: "d", Scope: LocalScope, Index: 1},
		"e": {Name: "e", Scope: LocalScope, Index: 0},
		"f": {Name: "f", Scope: LocalScope, Index: 1},
	}

	global := NewSymbolTable()
	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}

	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("expected a=%+v, got=%+v", expected["b"], b)
	}

	firstLocal := NewEnclosedSymbolTable(global)

	c := firstLocal.Define("c")
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}

	d := firstLocal.Define("d")
	if d != expected["d"] {
		t.Errorf("expected d=%+v, got=%+v", expected["d"], d)
	}

	secondLocal := NewEnclosedSymbolTable(firstLocal)

	e := secondLocal.Define("e")
	if e != expected["e"] {
		t.Errorf("expected e=%+v, got=%+v", expected["e"], e)
	}

	f := secondLocal.Define("f")
	if f != expected["f"] {
		t.Errorf("expected f=%+v, got=%+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	expected := []Symbol{
		Symbol{Name: "a", Scope: GlobalScope, Index: 0},
		Symbol{Name: "b", Scope: GlobalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	local := NewEnclosedSymbolTable(global)
	local.Define("c")
	local.Define("d")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "b", Scope: GlobalScope, Index: 1},
		{Name: "c", Scope: LocalScope, Index: 0},
		{Name: "d", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			table: firstLocal,
			expectedSymbols: []Symbol{
				{Name: "a", Scope: GlobalScope, Index: 0},
				{Name: "b", Scope: GlobalScope, Index: 1},
				{Name: "c", Scope: LocalScope, Index: 0},
				{Name: "d", Scope: LocalScope, Index: 1},
			},
		},
		{
			secondLocal,
			[]Symbol{
				Symbol{Name: "a", Scope: GlobalScope, Index: 0},
				Symbol{Name: "b", Scope: GlobalScope, Index: 1},
				Symbol{Name: "e", Scope: LocalScope, Index: 0},
				Symbol{Name: "f", Scope: LocalScope, Index: 1},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v",
					sym.Name, sym, result)
			}
		}
	}
}

func TestResolveBuiltins(t *testing.T) {
	global := NewSymbolTable()
	firstLocal := NewEnclosedSymbolTable(global)
	secondLocal := NewEnclosedSymbolTable(firstLocal)

	expected := []Symbol{
		{
			Name:  "a",
			Scope: BuiltinScope,
			Index: 0,
		},
		{
			Name:  "b",
			Scope: BuiltinScope,
			Index: 1,
		},
		{
			Name:  "c",
			Scope: BuiltinScope,
			Index: 2,
		},
		{
			Name:  "f",
			Scope: BuiltinScope,
			Index: 3,
		},
	}
	for i, v := range expected {
		global.DefineBuiltin(i, v.Name)
	}

	for _, table := range []*SymbolTable{
		global,
		firstLocal,
		secondLocal,
	} {
		for _, sym := range expected {
			result, ok := table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v", sym.Name, sym, result)
			}
		}
	}
}

// TestResolveFree sets up three scopes:
//   - firstLocal has c, d
//   - secondLocal has e, f
//     First, the expectedSymbols loop runs on secondLocal:
//     It calls secondLocal.Resolve("a") — found in global, returned as GlobalScope, no side effect.
//     It calls secondLocal.Resolve("b") — same, GlobalScope, no side effect.
//     It calls secondLocal.Resolve("c") — not in secondLocal.store, so goes to firstLocal.Resolve("c"), finds it as LocalScope.
//     Back in secondLocal.Resolve, since it's not Global or Builtin, calls secondLocal.defineFree({c, LocalScope, 0}).
//     This appends {c, LocalScope, 0} to secondLocal.FreeSymbols and stores {c, FreeScope, 0} in secondLocal.store.
//     Now secondLocal.FreeSymbols = [{c, LocalScope, 0}].
//     It calls secondLocal.Resolve("d") — same process, appends {d, LocalScope, 1} to FreeSymbols.
//     Now secondLocal.FreeSymbols = [{c, LocalScope, 0}, {d, LocalScope, 1}].
//     It calls secondLocal.Resolve("e") and "f" — found directly in secondLocal.store as LocalScope, no side effect.
//     Then the expectedFreeSymbols loop runs:
//     i=0: pulls secondLocal.FreeSymbols[0] which is {c, LocalScope, 0}, compares to expectedFreeSymbols[0] which is {c, LocalScope, 0}. Match.
//     i=1: pulls secondLocal.FreeSymbols[1] which is {d, LocalScope, 1}, compares to expectedFreeSymbols[1] which is {d, LocalScope, 1}. Match.
//     Test passes.
func TestResolveFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	tests := []struct {
		table               *SymbolTable
		expectedSymbols     []Symbol
		expectedFreeSymbols []Symbol
	}{
		{
			table: firstLocal,
			expectedSymbols: []Symbol{
				{
					Name:  "a",
					Scope: GlobalScope,
					Index: 0,
				},
				{
					Name:  "b",
					Scope: GlobalScope,
					Index: 1,
				},
				{
					Name:  "c",
					Scope: LocalScope,
					Index: 0,
				},
				{
					Name:  "d",
					Scope: LocalScope,
					Index: 1,
				},
			},
			expectedFreeSymbols: []Symbol{},
		},
		{
			table: secondLocal,
			expectedSymbols: []Symbol{
				{
					Name:  "a",
					Scope: GlobalScope,
					Index: 0,
				},
				{
					Name:  "b",
					Scope: GlobalScope,
					Index: 1,
				},
				{
					Name:  "c",
					Scope: FreeScope,
					Index: 0,
				},
				{
					Name:  "d",
					Scope: FreeScope,
					Index: 1,
				},
				{
					Name:  "e",
					Scope: LocalScope,
					Index: 0,
				},
				{
					Name:  "f",
					Scope: LocalScope,
					Index: 1,
				},
			},
			// which symbols play the role of free symbols
			expectedFreeSymbols: []Symbol{
				{
					Name:  "c",
					Scope: LocalScope, // describes where c originally lives
					Index: 0,
				},
				{
					Name:  "d",
					Scope: LocalScope, // describes where d originally lives
					Index: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v",
					sym.Name, sym, result)
			}
		}

		if len(tt.table.FreeSymbols) != len(tt.expectedFreeSymbols) {
			t.Errorf("wrong number of free symbols. got=%d, want=%d",
				len(tt.table.FreeSymbols), len(tt.expectedFreeSymbols))
			continue
		}

		for i, sym := range tt.expectedFreeSymbols {
			result := tt.table.FreeSymbols[i]
			if result != sym {
				t.Errorf("wrong free symbol. got=%+v, want=%+v",
					result, sym)
			}
		}
	}
}

func TestResolveUnresolvableFree(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")

	firstLocal := NewEnclosedSymbolTable(global)
	firstLocal.Define("c")

	secondLocal := NewEnclosedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")

	expected := []Symbol{
		{Name: "a", Scope: GlobalScope, Index: 0},
		{Name: "c", Scope: FreeScope, Index: 0},
		{Name: "e", Scope: LocalScope, Index: 0},
		{Name: "f", Scope: LocalScope, Index: 1},
	}

	for _, sym := range expected {
		result, ok := secondLocal.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}

	expectedUnresolvable := []string{
		"b",
		"d",
	}

	for _, name := range expectedUnresolvable {
		_, ok := secondLocal.Resolve(name)
		if ok {
			t.Errorf("name %s resolved, but was expected not to", name)
		}
	}
}

func TestDefineAndResolveFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")

	expected := Symbol{
		Name:  "a",
		Scope: FunctionScope,
		Index: 0,
	}
	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %s is not resolvable", expected.Name)
	}

	if result != expected {
		t.Errorf("Expected %s to resolve to %+v got %+v", expected.Name, expected, result)
	}
}

func TestShadowingFunctionName(t *testing.T) {
	global := NewSymbolTable()
	global.DefineFunctionName("a")
	global.Define("a")

	expected := Symbol{Name: "a", Scope: GlobalScope, Index: 0}

	result, ok := global.Resolve(expected.Name)
	if !ok {
		t.Fatalf("function name %s is not resolvable", expected.Name)
	}

	if result != expected {
		t.Errorf("Expected %s to resolve to %+v got %+v", expected.Name, expected, result)
	}
}
