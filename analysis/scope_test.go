package analysis

import (
	"testing"
)

func TestNewRootScope(t *testing.T) {
	rootSymbTable := newRootScope()

	for name, symbolType := range builtins {
		symbol, ok := rootSymbTable.Symbols[name]
		if !ok {
			t.Fatalf("Inbuilt symbol with name '%s' not found.", name)
		}

		if symbol.Type != symbolType {
			t.Errorf("Inbuilt symbol '%s': expected type %s, got %s.", name, symbolType, symbol.Type)
		}
	}
}

func TestDefineSymbol(t *testing.T) {
	rootSymbTable := newRootScope()
	rootSymbTable.define("test", variable, nil)

	t.Run("Existence", func(t *testing.T) {
		symbol, ok := rootSymbTable.Symbols["test"]
		if !ok || symbol == nil {
			t.Fatalf("Symbol 'test' not defined properly")
		}

		if symbol.Name != "test" {
			t.Errorf("Expected symbol name %s, got %s", "test", symbol.Name)
		}

		if symbol.Type != variable {
			t.Errorf("Expected symbol type %s, got %s", variable, symbol.Type)
		}
	})

	t.Run("Re-definition", func(t *testing.T) {
		ok := rootSymbTable.define("test", variable, nil)
		if !ok {
			t.Error("Expected ok to be true, got false")
		}

		symbol, _ := rootSymbTable.Symbols["test"]
		if !symbol.Used {
			t.Error("Expected Used to be true, got false")
		}
	})
}

func TestLookupSymbol(t *testing.T) {
	rootSymbTable := newRootScope()
	ok := rootSymbTable.define("var_1", variable, nil)
	if ok {
		t.Error("Expected ok to be false, got true")
	}

	t.Run("Symbol not found", func(t *testing.T) {
		symbol, ok := rootSymbTable.lookup("var_2")
		if ok {
			t.Error("Expected ok to be false, got true")
		}

		if symbol != nil {
			t.Errorf("Expected symbol have no value, got %#v", symbol)
		}
	})

	t.Run("Symbol found", func(t *testing.T) {
		symbol, ok := rootSymbTable.lookup("var_1")
		if !ok {
			t.Error("Expected ok to be true, got false")
		}

		if symbol == nil {
			t.Error("Expected symbol to have value, got nil")
		}
	})

	t.Run("Nested scope", func(t *testing.T) {
		child := newChildScope(rootSymbTable)
		grandChild := newChildScope(child)

		symbol, ok := grandChild.lookup("var_1")
		if !ok {
			t.Error("Expected ok to be true, got false")
		}

		if symbol == nil {
			t.Error("Expected symbol to have value, got nil")
		}
	})
}
