package loglint

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func messageExpr(pass *analysis.Pass, call *ast.CallExpr) (ast.Expr, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	if idx, ok := methodMsgIndex(pass, sel); ok {
		return argAt(call, idx)
	}
	if idx, ok := pkgFuncMsgIndex(pass, sel); ok {
		return argAt(call, idx)
	}
	return nil, false
}

func argAt(call *ast.CallExpr, idx int) (ast.Expr, bool) {
	if idx < 0 || idx >= len(call.Args) {
		return nil, false
	}
	return call.Args[idx], true
}

func methodMsgIndex(pass *analysis.Pass, sel *ast.SelectorExpr) (int, bool) {
	selection := pass.TypesInfo.Selections[sel]
	if selection == nil {
		return 0, false
	}
	recv := selection.Recv()
	name := sel.Sel.Name
	if isNamedType(recv, "log/slog", "Logger") {
		return slogMethodMsgIndex(name)
	}
	if isNamedType(recv, "go.uber.org/zap", "Logger") {
		return zapLoggerMsgIndex(name)
	}
	if isNamedType(recv, "go.uber.org/zap", "SugaredLogger") {
		return zapSugaredMsgIndex(name)
	}
	return 0, false
}

func pkgFuncMsgIndex(pass *analysis.Pass, sel *ast.SelectorExpr) (int, bool) {
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return 0, false
	}
	obj, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return 0, false
	}
	if obj.Imported().Path() == "log/slog" {
		return slogPkgMsgIndex(sel.Sel.Name)
	}
	return 0, false
}

func isNamedType(t types.Type, path, name string) bool {
	if t == nil {
		return false
	}
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg().Path() == path && obj.Name() == name
}

func slogPkgMsgIndex(name string) (int, bool) {
	return slogMsgIndex(name)
}

func slogMethodMsgIndex(name string) (int, bool) {
	return slogMsgIndex(name)
}

func zapLoggerMsgIndex(name string) (int, bool) {
	switch name {
	case "Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal":
		return 0, true
	default:
		return 0, false
	}
}

func zapSugaredMsgIndex(name string) (int, bool) {
	if idx, ok := zapLoggerMsgIndex(name); ok {
		return idx, true
	}
	if len(name) > 1 {
		base := name[:len(name)-1]
		suffix := name[len(name)-1]
		if suffix == 'f' || suffix == 'w' {
			if idx, ok := zapLoggerMsgIndex(base); ok {
				return idx, true
			}
		}
	}
	return 0, false
}

func slogMsgIndex(name string) (int, bool) {
	switch name {
	case "Debug", "Info", "Warn", "Error":
		return 0, true
	case "DebugContext", "InfoContext", "WarnContext", "ErrorContext":
		return 1, true
	case "Log", "LogAttrs":
		return 2, true
	default:
		return 0, false
	}
}
