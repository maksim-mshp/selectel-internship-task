package loglint

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"
)

type messageParts struct {
	Full       string
	FullPos    token.Pos
	HasFull    bool
	Leading    string
	LeadingPos token.Pos
	HasLeading bool
	Literals   []literalPart
}

type literalPart struct {
	Value string
	Pos   token.Pos
	End   token.Pos
}

func extractMessageParts(pass *analysis.Pass, expr ast.Expr) messageParts {
	parts := messageParts{}
	if full, pos, ok := constString(pass, expr); ok {
		parts.Full = full
		parts.FullPos = pos
		parts.HasFull = true
	}
	lits := make([]literalPart, 0, 2)
	collectStringLiterals(pass, expr, &parts, &lits)
	if len(lits) > 0 {
		parts.Literals = lits
	}
	return parts
}

func constString(pass *analysis.Pass, expr ast.Expr) (string, token.Pos, bool) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", token.NoPos, false
	}
	return constant.StringVal(tv.Value), expr.Pos(), true
}

func collectStringLiterals(pass *analysis.Pass, expr ast.Expr, parts *messageParts, lits *[]literalPart) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return
		}
		s, err := strconv.Unquote(e.Value)
		if err != nil {
			s = e.Value
		}
		*lits = append(*lits, literalPart{Value: s, Pos: e.Pos(), End: e.End()})
		if !parts.HasLeading {
			parts.HasLeading = true
			parts.Leading = s
			parts.LeadingPos = e.Pos()
		}
	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return
		}
		if !isStringType(pass.TypesInfo.Types[e].Type) {
			return
		}
		collectStringLiterals(pass, e.X, parts, lits)
		collectStringLiterals(pass, e.Y, parts, lits)
	case *ast.ParenExpr:
		collectStringLiterals(pass, e.X, parts, lits)
	}
}

func isStringType(t types.Type) bool {
	if t == nil {
		return false
	}
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Kind() == types.String
}
