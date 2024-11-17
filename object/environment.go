package object

type Environment struct {
	store map[string]Object
}

// NewEnvironment creates and returns a new Environment with an empty store.
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

// Get retrieves an Object from the Environment by its name, returning the Object and a boolean indicating its presence.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	return obj, ok
}

// Set assigns the given Object to the specified name in the Environment's store and returns the Object.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
}
