package main

import (
	"github.com/maksim-mshp/selectel-internship-task/internal/analyzer/loglint"

	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() {
	unitchecker.Main(loglint.Analyzer)
}
