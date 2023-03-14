package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"prom-metric-analyze/config"

	"github.com/sirupsen/logrus"
)

const (
	tmpDIR = "./.output/"

	metricPrefix = "metric"
	rulePrefix   = "rule"
)

func StartAnalyze(binary string) error {
	logrus.Warnln("start analyze")
	/*
		1. 通过 Grafana API 分析 Grafana 用到的指标
		mimirtool analyze grafana --address http://localhost:3000 --key=aaa
	*/
	os.Mkdir(tmpDIR, 0777)

	grafanaPath := urlToFileName(metricPrefix, config.Get().Grafana.RemoteURL)
	if err := fetchGrafanaAnalyze(binary, grafanaPath); err != nil {
		return err
	}

	// 2. 分析 Prometheus Alerting 和 Recording Rules 用到的指标
	rulePath := urlToFileName(rulePrefix, config.Get().Prometheus.RemoteURL)
	if err := fetchPrometheusRuleAnalyze(binary, rulePath); err != nil {
		return err
	}

	// 3. 分析指标
	analyzeMetrics(binary, grafanaPath, rulePath)

	// TODO 4. 生成优化建议，例如 write_relabel 配置或 metric_relabel 配置

	return nil
}

func urlToFileName(preFix string, url string) string {
	var app string
	switch preFix {
	case metricPrefix:
		app = "grafana"
	case rulePrefix:
		app = "prometheus"
	}

	gaFileName := strings.TrimPrefix(strings.TrimPrefix(url, "http://"), "https://")
	return path.Join(tmpDIR, fmt.Sprintf("%s_in_%s_%s.json", preFix, strings.ReplaceAll(gaFileName, ":", "_"), app))
}

func analyzeMetrics(binary, gPath, rPath string) error {
	/*
		bin/mimirtool analyze prometheus --address http://10.0.0.100:30090 --grafana-metrics-file ./output/metric_in_10.0.0.100_30030_grafana.json \
		--ruler-metrics-file ./output/rule_in_10.0.0.100_30090_prometheus.json
	*/
	g := config.Get().Prometheus
	command := []string{
		binary,
		"analyze", "prometheus",
		"--address", g.RemoteURL,
		"--grafana-metrics-file", gPath,
		"--ruler-metrics-file", rPath,
		"--output", path.Join(tmpDIR, ".analyze-result.json"),
	}
	return execCommand(command)
}

func fetchPrometheusRuleAnalyze(binary, prPath string) error {
	g := config.Get().Prometheus
	command := []string{
		binary,
		"analyze", "rule-file", g.LocalRuleFile,
		"--output", prPath,
	}
	return execCommand(command)
}

func fetchGrafanaAnalyze(binary, gaPath string) error {
	g := config.Get().Grafana
	command := []string{
		binary,
		"analyze", "grafana",
		"--address", g.RemoteURL,
		"--key", g.APIToken,
		"--output", gaPath,
	}

	return execCommand(command)
}

func execCommand(command []string) error {
	cmd := exec.Command("/bin/bash", "-c", strings.Join(command, " "))
	logrus.Warnln("analyze command:", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorln("cmd.CombinedOutput() failed", err, string(out))
		return err
	}

	if len(out) > 0 {
		logrus.Warnln("mimirtool output:", string(out))
	}

	return nil
}
