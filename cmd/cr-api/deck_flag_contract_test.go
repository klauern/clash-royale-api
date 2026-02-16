package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestDeckAnalyzeFlagContract(t *testing.T) {
	assertRuntimeFlagsDeclared(
		t,
		"deck_optimize_runtime.go",
		"deckAnalyzeCommand",
		addDeckAnalyzeCommand(),
	)
}

func TestDeckOptimizeFlagContract(t *testing.T) {
	assertRuntimeFlagsDeclared(
		t,
		"deck_optimize_runtime.go",
		"deckOptimizeCommand",
		addDeckOptimizeCommand(),
	)
}

func TestDeckRecommendFlagContract(t *testing.T) {
	assertRuntimeFlagsDeclared(
		t,
		"deck_recommend_mulligan_runtime.go",
		"deckRecommendCommand",
		addDeckRecommendCommand(),
	)
}

func assertRuntimeFlagsDeclared(t *testing.T, fileName, funcName string, command *cli.Command) {
	t.Helper()

	declared := commandFlagSet(command)
	for globalFlag := range globalFlagSet() {
		declared[globalFlag] = struct{}{}
	}

	readFlags, err := runtimeReadFlags(filepath.Join(".", fileName), funcName)
	if err != nil {
		t.Fatalf("failed to inspect runtime flags: %v", err)
	}

	for flag := range readFlags {
		if _, ok := declared[flag]; !ok {
			t.Fatalf("runtime reads undeclared flag %q in %s", flag, funcName)
		}
	}
}

func commandFlagSet(command *cli.Command) map[string]struct{} {
	result := make(map[string]struct{})
	if command == nil {
		return result
	}

	for _, flag := range command.Flags {
		for _, name := range flag.Names() {
			result[name] = struct{}{}
		}
	}
	return result
}

func globalFlagSet() map[string]struct{} {
	return map[string]struct{}{
		"api-token": {},
		"data-dir":  {},
		"verbose":   {},
	}
}

func runtimeReadFlags(filePath, funcName string) (map[string]struct{}, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		return nil, err
	}

	flagMethods := map[string]struct{}{
		"String":      {},
		"StringSlice": {},
		"Int":         {},
		"Bool":        {},
		"Float64":     {},
	}
	result := make(map[string]struct{})

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || fn.Name.Name != funcName || fn.Body == nil {
			continue
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok || ident.Name != "cmd" {
				return true
			}
			if _, ok := flagMethods[selector.Sel.Name]; !ok {
				return true
			}
			if len(call.Args) == 0 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			flagName := lit.Value
			if len(flagName) >= 2 {
				flagName = flagName[1 : len(flagName)-1]
			}
			result[flagName] = struct{}{}
			return true
		})
	}

	return result, nil
}
