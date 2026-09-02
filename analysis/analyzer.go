package analysis

import (
	"fmt"

	"github.com/SegniAT/monkey-language-interpreter/ast"
	"github.com/SegniAT/monkey-language-interpreter/token"
)

// analyze builds the document's symbol table, resolves every identifier into
// the Uses map, records every function scope in the Scopes map and returns
// semantic diagnostics for undefined, redeclared and unused names.
func (d *Document) analyze() []token.Diagnostic {
	d.rootScope = newRootScope()
	d.identifierUses = make(map[*ast.Identifier]*symbol)
	d.functionScopes = make(map[*ast.FunctionLiteral]*symbolTable)

	var diags []token.Diagnostic
	d.visit(d.AST, d.rootScope, &diags)
	diags = append(diags, d.unusedDiagnostics()...)

	return diags
}

func (d *Document) visit(node ast.Node, scope *symbolTable, diags *[]token.Diagnostic) {
	if node == nil {
		return
	}

	switch node := node.(type) {
	case *ast.Program:
		for _, statement := range node.Statements {
			d.visit(statement, scope, diags)
		}

	case *ast.BlockStatement:
		for _, statement := range node.Statements {
			d.visit(statement, scope, diags)
		}

	case *ast.LetStatement:
		if scope.define(node.Name.Value, variable, node.Name) {
			*diags = append(*diags, token.Diagnostic{
				Message:  fmt.Sprintf("redeclaration of %s", node.Name.Value),
				Range:    node.Name.Token.Range,
				Severity: token.Error,
			})
		}
		d.visit(node.Value, scope, diags)

	case *ast.FunctionLiteral:
		child := newChildScope(scope)
		d.functionScopes[node] = child
		for _, param := range node.Parameters {
			if child.define(param.Value, parameter, param) {
				*diags = append(*diags, token.Diagnostic{
					Message:  fmt.Sprintf("redeclaration of %s", param.Value),
					Range:    param.Token.Range,
					Severity: token.Error,
				})
			}
		}
		d.visit(node.Body, child, diags)

	case *ast.ExpressionStatement:
		d.visit(node.Expression, scope, diags)

	case *ast.ReturnStatement:
		d.visit(node.ReturnValue, scope, diags)

	case *ast.IfExpression:
		d.visit(node.Condition, scope, diags)
		d.visit(node.Consequence, scope, diags)
		if node.Alternative != nil {
			d.visit(node.Alternative, scope, diags)
		}

	case *ast.PrefixExpression:
		d.visit(node.Right, scope, diags)

	case *ast.InfixExpression:
		d.visit(node.Left, scope, diags)
		d.visit(node.Right, scope, diags)

	case *ast.CallExpression:
		d.visit(node.Function, scope, diags)
		for _, arg := range node.Arguments {
			d.visit(arg, scope, diags)
		}

	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			d.visit(element, scope, diags)
		}

	case *ast.IndexExpression:
		d.visit(node.Left, scope, diags)
		d.visit(node.Index, scope, diags)

	case *ast.HashLiteral:
		for key, value := range node.Pairs {
			d.visit(key, scope, diags)
			d.visit(value, scope, diags)
		}

	case *ast.Identifier:
		if sym, ok := scope.lookup(node.Value); ok {
			sym.Used = true
			d.identifierUses[node] = sym
		} else {
			*diags = append(*diags, token.Diagnostic{
				Message:  fmt.Sprintf("undefined variable: %s", node.Value),
				Range:    node.Token.Range,
				Severity: token.Error,
			})
		}

		// *ast.IntegerLiteral, *ast.Boolean and *ast.StringLiteral are leaves.
	}
}

// unusedDiagnostics reports user-declared variables and functions that were
// never resolved as a use anywhere in the document.
func (d *Document) unusedDiagnostics() []token.Diagnostic {
	var diags []token.Diagnostic

	scopes := []*symbolTable{d.rootScope}
	for _, child := range d.functionScopes {
		scopes = append(scopes, child)
	}

	for _, scope := range scopes {
		for name, symbol := range scope.Symbols {
			if symbol.Used || symbol.Identifier == nil {
				continue
			}
			if symbol.Type != variable && symbol.Type != parameter {
				continue
			}

			if symbol.Name == "_" {
				continue
			}

			diags = append(diags, token.Diagnostic{
				Message: func() string {
					switch symbol.Type {
					case variable:
						return fmt.Sprintf("unused variable: %s", name)
					case parameter:
						return fmt.Sprintf("unused parameter: %s", name)
					default:
						return "unknown message"
					}
				}(),
				Range:    token.Range{Start: symbol.Identifier.Start(), End: symbol.Identifier.End()},
				Severity: token.Warning,
			})
		}
	}

	return diags
}
