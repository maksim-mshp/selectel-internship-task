package loglint

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

func reportViolations(pass *analysis.Pass, msgExpr ast.Expr, parts messageParts) {
	if parts.HasFull {
		reportAll(pass, parts.FullPos, parts.Full, true)
		return
	}
	if !parts.HasLeading && len(parts.Literals) == 0 {
		return
	}
	if parts.HasLeading && violatesLowercase(parts.Leading) {
		pass.Reportf(parts.LeadingPos, "log message must start with a lowercase letter")
	}
	var english, special, sensitive bool
	for _, lit := range parts.Literals {
		if !english && containsNonEnglishLetter(lit.Value) {
			english = true
		}
		if !special && containsSpecialOrEmoji(lit.Value) {
			special = true
		}
		if !sensitive && containsSensitiveKeyword(lit.Value) {
			sensitive = true
		}
	}
	pos := msgExpr.Pos()
	if len(parts.Literals) > 0 {
		pos = parts.Literals[0].Pos
	}
	if english {
		pass.Reportf(pos, "log message must be in English")
	}
	if special {
		pass.Reportf(pos, "log message must not contain special symbols or emoji")
	}
	if sensitive {
		pass.Reportf(pos, "log message must not contain sensitive data")
	}
}

func reportAll(pass *analysis.Pass, pos token.Pos, msg string, includeLower bool) {
	if includeLower && violatesLowercase(msg) {
		pass.Reportf(pos, "log message must start with a lowercase letter")
	}
	if containsNonEnglishLetter(msg) {
		pass.Reportf(pos, "log message must be in English")
	}
	if containsSpecialOrEmoji(msg) {
		pass.Reportf(pos, "log message must not contain special symbols or emoji")
	}
	if containsSensitiveKeyword(msg) {
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

func containsSensitiveKeyword(s string) bool {
	low := strings.ToLower(s)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(low, kw) {
			return true
		}
	}
	return false
}

var sensitiveKeywords = []string{
	"password",
	"passwd",
	"pwd",
	"secret",
	"token",
	"access_token",
	"refresh_token",
	"id_token",
	"api_key",
	"apikey",
	"access_key",
	"private_key",
	"client_secret",
	"authorization",
	"bearer",
	"session",
	"cookie",
	"ssn",
	"cvv",
	"pin",
}
