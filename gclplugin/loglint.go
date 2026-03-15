package gclplugin

import (
	"github.com/maksim-mshp/selectel-internship-task/internal/analyzer/loglint"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("loglint", New)
}

type Plugin struct {
	cfg loglint.Config
}

type Settings struct {
	Config string `json:"config" yaml:"config"`
}

func New(conf any) (register.LinterPlugin, error) {
	settings, err := register.DecodeSettings[Settings](conf)
	if err != nil {
		return nil, err
	}
	cfg, err := loglint.LoadConfig(settings.Config)
	if err != nil {
		return nil, err
	}
	return &Plugin{cfg: cfg}, nil
}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{loglint.NewAnalyzer(p.cfg)}, nil
}

func (p *Plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
