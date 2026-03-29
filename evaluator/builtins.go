package evaluator

import (
	"monkey/object"
)

var builtins = map[string]*object.Builtin{
	"len":   object.GetBuiltinByName("len"),
	"first": object.GetBuiltinByName("first"),
	"last":  object.GetBuiltinByName("last"),
	"tail":  object.GetBuiltinByName("tail"),
	"push":  object.GetBuiltinByName("push"),
	"puts":  object.GetBuiltinByName("puts"),
	"rest":  object.GetBuiltinByName("rest"),
}
