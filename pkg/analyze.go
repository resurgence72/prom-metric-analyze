package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"prom-metric-analyze/config"

	"github.com/sirupsen/logrus"
)

const (
	tmpDIR = "./.output/"

	metricPrefix = "metric"
	rulePrefix   = "rule"
)

type metricAnalyze struct {
	InUseMetrics []*useMetric `json:"in_use_metric_counts"`
}

type useMetric struct {
	Metric string `json:"metric"`
	Count  int64  `json:"count"`
}

func StartAnalyze(binary string) error {
	logrus.Warnln("start analyze")
	/*
		1. 通过 Grafana API 分析 Grafana 用到的指标
		mimirtool analyze grafana --address http://localhost:3000 --key=aaa
	*/
	os.RemoveAll(tmpDIR)
	os.Mkdir(tmpDIR, 0755)

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
	metrics, err := analyzeMetrics(binary, grafanaPath, rulePath)
	if err != nil {
		return err
	}

	// 4. 生成优化建议，例如 write_relabel 配置或 metric_relabel 配置
	generatingOptimalConfig(metrics)
	return nil
}

func generatingOptimalConfig(metrics []string) {
	metricType := make(map[string][]string)
	compile := regexp.MustCompile("^([a-zA-Z0-9]+?)[:_].*")

	getMetricPrefix := func(m string) string {
		findString := compile.FindStringSubmatch(m)
		if len(findString) < 2 {
			return "others"
		}

		return findString[len(findString)-1]
	}

	for _, metric := range metrics {
		mPre := getMetricPrefix(metric)
		if _, ok := metricType[mPre]; !ok {
			metricType[mPre] = []string{metric}
		} else {
			metricType[mPre] = append(metricType[mPre], metric)
		}
	}

	// 根据前缀生成relabel文件
	optimizationDIR := "./metrics_analyze_result"
	os.RemoveAll(optimizationDIR)
	os.Mkdir(optimizationDIR, 0755)

	var (
		buf     bytes.Buffer
		relabel = config.Get().OptimizationRelabelType
	)

	for k := range metricType {
		buf.WriteString(relabel)
		buf.WriteString(":\n")
		buf.WriteString("- action: keep\n")
		buf.WriteString("  source_labels: [ __name__ ]\n")
		buf.WriteString("  regex: ")
		buf.WriteString(strings.Join(metricType[k], "|"))
		buf.WriteString("\n")

		f := path.Join(optimizationDIR, fmt.Sprintf("suggest_for_%s_with_prefix_%s", relabel, k))
		if err := os.WriteFile(f, buf.Bytes(), 0755); err != nil {
			logrus.Errorln("write analyze result failed", f, err)
		}

		buf.Reset()
	}

	logrus.Warnf("The optimized configuration segment has been generated in the %s directory\n", optimizationDIR)
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

func analyzeMetrics(binary, gPath, rPath string) ([]string, error) {
	/*
		bin/mimirtool analyze prometheus --address http://10.0.0.100:30090 --grafana-metrics-file ./output/metric_in_10.0.0.100_30030_grafana.json \
		--ruler-metrics-file ./output/rule_in_10.0.0.100_30090_prometheus.json
	*/
	outputFile := ".analyze-result.json"
	outputPath := path.Join(tmpDIR, outputFile)

	g := config.Get().Prometheus
	command := []string{
		binary,
		"analyze", "prometheus",
		"--address", g.RemoteURL,
		"--grafana-metrics-file", gPath,
		"--ruler-metrics-file", rPath,
		"--output", outputPath,
	}
	if err := execCommand(command); err != nil {
		return nil, err
	}

	bs, err := os.ReadFile(outputPath)
	if err != nil {
		logrus.Errorln("readFile output file failed", err)
		return nil, err
	}

	ma := new(metricAnalyze)
	if err = json.Unmarshal(bs, ma); err != nil {
		logrus.Errorln("json.Unmarshal output file failed", err)
		return nil, err
	}

	var metrics []string
	for _, m := range ma.InUseMetrics {
		m := m
		metrics = append(metrics, m.Metric)
	}
	return metrics, nil
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
		logrus.Warnf("mimirtool output:%s\n\n", string(out))
	}

	return nil
}
