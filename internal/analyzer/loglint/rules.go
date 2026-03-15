package loglint

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

func reportViolations(pass *analysis.Pass, msgExpr ast.Expr, parts messageParts, rules ruleSet) {
	if parts.HasFull {
		reportAll(pass, parts.FullPos, parts.Full, rules)
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
		if rules.sensitive && !sensitive && containsSensitiveKeyword(lit.Value, rules.sensitiveKeywords) {
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
		pass.Reportf(pos, "log message must not contain special symbols or emoji")
	}
	if rules.sensitive && sensitive {
		pass.Reportf(pos, "log message must not contain sensitive data")
	}
}

func reportAll(pass *analysis.Pass, pos token.Pos, msg string, rules ruleSet) {
	if rules.lowercase && violatesLowercase(msg) {
		pass.Reportf(pos, "log message must start with a lowercase letter")
	}
	if rules.english && containsNonEnglishLetter(msg) {
		pass.Reportf(pos, "log message must be in English")
	}
	if rules.special && containsSpecialOrEmoji(msg) {
		pass.Reportf(pos, "log message must not contain special symbols or emoji")
	}
	if rules.sensitive && containsSensitiveKeyword(msg, rules.sensitiveKeywords) {
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

func containsSensitiveKeyword(s string, keywords []string) bool {
	low := strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(low, kw) {
			return true
		}
	}
	return false
}
