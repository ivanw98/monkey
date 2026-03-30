package compiler

// SymbolScope identifies the scope in which a symbol is defined.
type SymbolScope string

const (
	// LocalScope marks a symbol as local to an enclosed scope.
	LocalScope SymbolScope = "LOCAL"
	// GlobalScope marks a symbol as defined in the global scope.
	GlobalScope SymbolScope = "GLOBAL"
	// BuiltinScope marks a symbol as defined in the builtin scope.
	BuiltinScope SymbolScope = "BUILTIN"
)

// Symbol describes a named binding tracked by the compiler.
type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

// SymbolTable acts like a linked chain of nested scopes.
type SymbolTable struct {
	Outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
}

// Define creates a symbol for name in the current table and returns it.
func (st *SymbolTable) Define(name string) Symbol {
	sym := Symbol{
		Name:  name,
		Index: st.numDefinitions,
		Scope: LocalScope,
	}
	// numDefinitions is correct for both globals and locals because it counts definitions within the current symbol table,
	// and each symbol table represents exactly one scope.

	if st.Outer == nil {
		sym.Scope = GlobalScope
	}

	st.store[name] = sym
	st.numDefinitions++
	return sym
}

// Resolve looks up name in the current table, it recursively checks Outer tables if not found.
func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := st.store[name]
	if !ok && st.Outer != nil {
		return st.Outer.Resolve(name)
	}
	return obj, ok
}

func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	sym := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	st.store[name] = sym
	return sym
}

// NewSymbolTable creates a new top-level symbol table.
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	return &SymbolTable{
		store: s,
	}
}

// NewEnclosedSymbolTable creates a new symbol table enclosed by outer.
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.Outer = outer
	return s
}
