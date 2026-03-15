package main

import (
	"log-linter/internal/analyzer/loglint"

	"golang.org/x/tools/go/analysis"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{loglint.Analyzer}, nil
}
