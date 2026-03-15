package loglint

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/packages"
)

func TestLowercase(t *testing.T) {
	diags := runOnPackage(t, LowercaseAnalyzer, "lowercase")
	requireSingleMessage(t, diags, "log message must start with a lowercase letter")
}

func TestEnglish(t *testing.T) {
	diags := runOnPackage(t, EnglishAnalyzer, "english")
	requireSingleMessage(t, diags, "log message must be in English")
}

func TestSpecial(t *testing.T) {
	diags := runOnPackage(t, SpecialAnalyzer, "special")
	requireSingleMessage(t, diags, "log message must not contain special symbols or emoji")
}

func TestSensitive(t *testing.T) {
	diags := runOnPackage(t, SensitiveAnalyzer, "sensitive")
	requireSingleMessage(t, diags, "log message must not contain sensitive data")
}

func runOnPackage(t *testing.T, analyzer *analysis.Analyzer, pattern string) []analysis.Diagnostic {
	t.Helper()
	root := testdataRoot(t)
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes |
			packages.NeedImports |
			packages.NeedModule,
		Dir: root,
		Env: append(os.Environ(), "GOPATH="+root, "GO111MODULE=off"),
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		t.Fatalf("load package: %v", err)
	}
	if len(pkgs) != 1 {
		t.Fatalf("expected one package, got %d", len(pkgs))
	}
	pkg := pkgs[0]
	if len(pkg.Errors) > 0 {
		t.Fatalf("package errors: %v", pkg.Errors[0])
	}
	inspectPass := &analysis.Pass{
		Analyzer:   inspect.Analyzer,
		Fset:       pkg.Fset,
		Files:      pkg.Syntax,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
		Report:     func(analysis.Diagnostic) {},
	}
	inspectResult, err := inspect.Analyzer.Run(inspectPass)
	if err != nil {
		t.Fatalf("inspect run: %v", err)
	}
	var diags []analysis.Diagnostic
	pass := &analysis.Pass{
		Analyzer:   analyzer,
		Fset:       pkg.Fset,
		Files:      pkg.Syntax,
		Pkg:        pkg.Types,
		TypesInfo:  pkg.TypesInfo,
		TypesSizes: pkg.TypesSizes,
		ResultOf: map[*analysis.Analyzer]interface{}{
			inspect.Analyzer: inspectResult,
		},
		Report: func(d analysis.Diagnostic) {
			diags = append(diags, d)
		},
	}
	_, err = analyzer.Run(pass)
	if err != nil {
		t.Fatalf("analyzer run: %v", err)
	}
	return diags
}

func requireSingleMessage(t *testing.T, diags []analysis.Diagnostic, msg string) {
	t.Helper()
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Message != msg {
		t.Fatalf("unexpected diagnostic: %s", diags[0].Message)
	}
}

func testdataRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	return filepath.Clean(filepath.Join(wd, "testdata"))
}
