package main

import (
	"flag"
	"os"

	"prom-metric-analyze/pkg"
	"prom-metric-analyze/pkg/config"

	"github.com/sirupsen/logrus"
)

func main() {
	var (
		configPath string
		h          bool
	)
	flag.StringVar(&configPath, "config.file", "./analyze.yaml", "config file path")
	flag.BoolVar(&h, "h", false, "help")
	flag.Parse()

	if h {
		flag.Usage()
		os.Exit(0)
	}

	if err := config.InitConfig(configPath); err != nil {
		logrus.Fatal(err)
	}

	// 检查mimirtool是否可用
	binaryPath, err := pkg.CheckMimirToolBinary()
	if err != nil {
		logrus.Fatal(err)
	}

	pkg.StartAnalyze(*binaryPath)
}
