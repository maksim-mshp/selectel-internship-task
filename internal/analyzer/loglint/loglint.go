package loglint

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "loglint",
	Doc:      "checks log messages",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var LowercaseAnalyzer = &analysis.Analyzer{
	Name:     "loglintlowercase",
	Doc:      "checks log message lowercase start",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runLowercase,
}

var EnglishAnalyzer = &analysis.Analyzer{
	Name:     "loglintenglish",
	Doc:      "checks log message language",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runEnglish,
}

var SpecialAnalyzer = &analysis.Analyzer{
	Name:     "loglintspecial",
	Doc:      "checks log message symbols",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runSpecial,
}

var SensitiveAnalyzer = &analysis.Analyzer{
	Name:     "loglintsensitive",
	Doc:      "checks log message sensitive data",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runSensitive,
}

type ruleSet struct {
	lowercase bool
	english   bool
	special   bool
	sensitive bool
}

func run(pass *analysis.Pass) (interface{}, error) {
	return runWith(pass, ruleSet{
		lowercase: true,
		english:   true,
		special:   true,
		sensitive: true,
	})
}

func runLowercase(pass *analysis.Pass) (interface{}, error) {
	return runWith(pass, ruleSet{lowercase: true})
}

func runEnglish(pass *analysis.Pass) (interface{}, error) {
	return runWith(pass, ruleSet{english: true})
}

func runSpecial(pass *analysis.Pass) (interface{}, error) {
	return runWith(pass, ruleSet{special: true})
}

func runSensitive(pass *analysis.Pass) (interface{}, error) {
	return runWith(pass, ruleSet{sensitive: true})
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
