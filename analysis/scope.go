package analysis

import (
	"github.com/SegniAT/monkey-language-interpreter/ast"
)

type symbolType string

const (
	variable  symbolType = "variable"
	parameter symbolType = "parameter"
	builtin   symbolType = "builtin"
)

type symbol struct {
	Name       string
	Type       symbolType
	Range      Range
	Identifier *ast.Identifier
	Used       bool
}

var builtins = []string{"len", "first", "last", "rest", "push", "puts"}

type symbolTable struct {
	Outer   *symbolTable
	Symbols map[string]*symbol
}

func newChildScope(outer *symbolTable) *symbolTable {
	return &symbolTable{
		Outer:   outer,
		Symbols: map[string]*symbol{},
	}
}

// newRootScope creates the global scope with the monkey builtins pre-registered
func newRootScope() *symbolTable {
	scope := &symbolTable{
		Symbols: map[string]*symbol{},
	}

	for _, name := range builtins {
		symbol := &symbol{
			Name: name,
			Type: builtin,
		}

		scope.Symbols[name] = symbol
	}

	return scope
}

// define registers a symbol in the scope and returns true if the name was
// already declared in this same scope (except builtins, which may be shadowed).
func (s *symbolTable) define(name string, kind symbolType, identifier *ast.Identifier) bool {
	redeclared := false
	if existing, ok := s.Symbols[name]; ok && existing.Type != builtin {
		redeclared = true
	}

	symbol := &symbol{
		Name: name,
		Type: kind,
		Range: func() Range {
			if identifier == nil {
				return Range{}
			}
			return Range{
				Start: Position{Line: identifier.Start().Line, Character: identifier.Start().Character},
				End:   Position{Line: identifier.End().Line, Character: identifier.End().Character},
			}
		}(),
		Identifier: identifier,
		Used:       redeclared,
	}

	s.Symbols[name] = symbol

	return redeclared
}

func (s *symbolTable) lookup(name string) (*symbol, bool) {
	sym, ok := s.Symbols[name]
	if ok {
		return sym, true
	}

	if s.Outer != nil {
		return s.Outer.lookup(name)
	}

	return nil, false
}
