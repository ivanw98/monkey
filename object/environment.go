package object

// Environment represents a storage for objects, maintaining a mapping between variable names and their corresponding objects.
type Environment struct {
	store map[string]Object
	outer *Environment
}

// NewEnclosedEnvironment creates a new Environment containing a reference to an outer Environment for nested scopes.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

// NewEnvironment creates and returns a new Environment with an empty store.
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{
		store: s,
	}
}

// Get retrieves an Object from the Environment by its name, returning the Object and a boolean indicating its presence.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set assigns the given Object to the specified name in the Environment's store and returns the Object.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
