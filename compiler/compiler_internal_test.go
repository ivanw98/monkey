package compiler

import (
	"monkey/code"
	"testing"
)

func TestCompilerScopes(t *testing.T) {
	comp := New()
	if comp.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 0)
	}

	comp.emit(code.OpMul)

	comp.enterScope()
	if comp.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", comp.scopeIndex, 1)
	}

	comp.emit(code.OpSub)

	if len(comp.scopes[comp.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d",
			len(comp.scopes[comp.scopeIndex].instructions))
	}

	last := comp.scopes[comp.scopeIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpSub)
	}

	comp.leaveScope()
	if comp.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d",
			comp.scopeIndex, 0)
	}

	comp.emit(code.OpAdd)

	if len(comp.scopes[comp.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d",
			len(comp.scopes[comp.scopeIndex].instructions))
	}

	last = comp.scopes[comp.scopeIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpAdd)
	}

	previous := comp.scopes[comp.scopeIndex].previousInstruction
	if previous.Opcode != code.OpMul {
		t.Errorf("previousInstruction.Opcode wrong. got=%d, want=%d",
			previous.Opcode, code.OpMul)
	}
}
