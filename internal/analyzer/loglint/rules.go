package loglint

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

func reportViolations(pass *analysis.Pass, msgExpr ast.Expr, parts messageParts, rules ruleSet) {
	if parts.HasFull {
		if rules.lowercase && violatesLowercase(parts.Full) {
			pass.Reportf(parts.FullPos, "log message must start with a lowercase letter")
		}
		if rules.english && containsNonEnglishLetter(parts.Full) {
			pass.Reportf(parts.FullPos, "log message must be in English")
		}
		if rules.special && containsSpecialOrEmoji(parts.Full) {
			diag := analysis.Diagnostic{
				Pos:     parts.FullPos,
				Message: "log message must not contain special symbols or emoji",
			}
			if fixes := buildSpecialFixes(msgExpr, parts); len(fixes) > 0 {
				diag.SuggestedFixes = fixes
			}
			pass.Report(diag)
		}
		if rules.sensitive && containsSensitivePattern(parts.Full, rules.sensitivePatterns) {
			pass.Reportf(parts.FullPos, "log message must not contain sensitive data")
		}
		return
	}
	if !parts.HasLeading && len(parts.Literals) == 0 {
		return
	}
	if rules.lowercase && parts.HasLeading && violatesLowercase(parts.Leading) {
		pass.Reportf(parts.LeadingPos, "log message must start with a lowercase letter")
	}
	var english, special, sensitive bool
	for _, lit := range parts.Literals {
		if rules.english && !english && containsNonEnglishLetter(lit.Value) {
			english = true
		}
		if rules.special && !special && containsSpecialOrEmoji(lit.Value) {
			special = true
		}
		if rules.sensitive && !sensitive && containsSensitivePattern(lit.Value, rules.sensitivePatterns) {
			sensitive = true
		}
	}
	pos := msgExpr.Pos()
	if len(parts.Literals) > 0 {
		pos = parts.Literals[0].Pos
	}
	if rules.english && english {
		pass.Reportf(pos, "log message must be in English")
	}
	if rules.special && special {
		diag := analysis.Diagnostic{
			Pos:     pos,
			Message: "log message must not contain special symbols or emoji",
		}
		if fixes := buildSpecialFixes(msgExpr, parts); len(fixes) > 0 {
			diag.SuggestedFixes = fixes
		}
		pass.Report(diag)
	}
	if rules.sensitive && sensitive {
		pass.Reportf(pos, "log message must not contain sensitive data")
	}
}

func violatesLowercase(s string) bool {
	trimmed := strings.TrimLeftFunc(s, unicode.IsSpace)
	if trimmed == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(trimmed)
	return !unicode.IsLower(r)
}

func containsNonEnglishLetter(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				continue
			}
			return true
		}
	}
	return false
}

func containsSpecialOrEmoji(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			continue
		}
		return true
	}
	return false
}

func containsSensitivePattern(s string, patterns []*regexp.Regexp) bool {
	for _, p := range patterns {
		if p.MatchString(s) {
			return true
		}
	}
	return false
}

func buildSpecialFixes(msgExpr ast.Expr, parts messageParts) []analysis.SuggestedFix {
	if lit, ok := msgExpr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		original, err := strconv.Unquote(lit.Value)
		if err != nil {
			original = lit.Value
		}
		cleaned := removeSpecialSymbols(original)
		if cleaned == "" || cleaned == original {
			return nil
		}
		newLit := strconv.Quote(cleaned)
		return []analysis.SuggestedFix{
			{
				Message: "remove special symbols",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     lit.Pos(),
						End:     lit.End(),
						NewText: []byte(newLit),
					},
				},
			},
		}
	}
	if len(parts.Literals) == 0 {
		return nil
	}
	if parts.HasFull {
		cleanedFull := removeSpecialSymbols(parts.Full)
		if cleanedFull == "" {
			return nil
		}
	}
	edits := make([]analysis.TextEdit, 0, len(parts.Literals))
	for _, lit := range parts.Literals {
		cleaned := removeSpecialSymbols(lit.Value)
		if cleaned != lit.Value {
			newLit := strconv.Quote(cleaned)
			edits = append(edits, analysis.TextEdit{
				Pos:     lit.Pos,
				End:     lit.End,
				NewText: []byte(newLit),
			})
		}
	}
	if len(edits) == 0 {
		return nil
	}
	return []analysis.SuggestedFix{
		{
			Message:   "remove special symbols",
			TextEdits: edits,
		},
	}
}

func removeSpecialSymbols(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
