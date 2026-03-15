package loglint

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = NewAnalyzer(DefaultConfig())

var LowercaseAnalyzer = NewRuleAnalyzer(
	"loglintlowercase",
	"checks log message lowercase start",
	ruleSet{lowercase: true},
	DefaultConfig(),
)

var EnglishAnalyzer = NewRuleAnalyzer(
	"loglintenglish",
	"checks log message language",
	ruleSet{english: true},
	DefaultConfig(),
)

var SpecialAnalyzer = NewRuleAnalyzer(
	"loglintspecial",
	"checks log message symbols",
	ruleSet{special: true},
	DefaultConfig(),
)

var SensitiveAnalyzer = NewRuleAnalyzer(
	"loglintsensitive",
	"checks log message sensitive data",
	ruleSet{sensitive: true},
	DefaultConfig(),
)

type ruleSet struct {
	lowercase         bool
	english           bool
	special           bool
	sensitive         bool
	sensitivePatterns []*regexp.Regexp
}

func NewAnalyzer(cfg Config) *analysis.Analyzer {
	return NewRuleAnalyzer("loglint", "checks log messages", rulesFromConfig(cfg), cfg)
}

func NewRuleAnalyzer(name, doc string, mask ruleSet, cfg Config) *analysis.Analyzer {
	base := rulesFromConfig(cfg)
	rules := ruleSet{
		lowercase:         mask.lowercase && base.lowercase,
		english:           mask.english && base.english,
		special:           mask.special && base.special,
		sensitive:         mask.sensitive && base.sensitive,
		sensitivePatterns: base.sensitivePatterns,
	}
	return newAnalyzerWithRules(name, doc, rules)
}

func rulesFromConfig(cfg Config) ruleSet {
	patterns := cfg.compiledPatterns
	if patterns == nil && cfg.Patterns != nil {
		compiled, _ := compilePatterns(cfg.Patterns)
		patterns = compiled
	}
	return ruleSet{
		lowercase:         cfg.Lowercase,
		english:           cfg.English,
		special:           cfg.Special,
		sensitive:         cfg.Sensitive,
		sensitivePatterns: patterns,
	}
}

func newAnalyzerWithRules(name, doc string, rules ruleSet) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     name,
		Doc:      doc,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return runWith(pass, rules)
		},
	}
}

func runWith(pass *analysis.Pass, rules ruleSet) (interface{}, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		msgExpr, ok := messageExpr(pass, call)
		if !ok {
			return
		}
		parts := extractMessageParts(pass, msgExpr)
		reportViolations(pass, msgExpr, parts, rules)
	})
	return nil, nil
}
