package config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type MetricAnalyze struct {
	Grafana      *grafana    `yaml:"grafana"`
	Prometheus   *prometheus `yaml:"prometheus"`
	MimirtoolDIR string      `yaml:"mimirtool_dir"`

	m sync.Mutex `yaml:"m"`
}

type prometheus struct {
	RemoteURL     string `yaml:"remote_url"`
	LocalRuleFile string `yaml:"local_rule_file"`
}

type grafana struct {
	RemoteURL string `yaml:"remote_url"`
	APIToken  string `yaml:"api_token"`
}

var cfg *MetricAnalyze

func InitConfig(path string) error {
	config, err := loadFile(path)
	if err != nil {
		return err
	}
	cfg = config
	return nil
}

func Get() *MetricAnalyze {
	cfg.m.Lock()
	defer cfg.m.Unlock()
	return cfg
}

func loadFile(fileName string) (*MetricAnalyze, error) {
	bytes, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return load(bytes)
}

func load(bytes []byte) (*MetricAnalyze, error) {
	cfg := &MetricAnalyze{}
	err := yaml.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
