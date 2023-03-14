package main

import (
	"flag"
	"os"

	"prom-metric-analyze/config"
	"prom-metric-analyze/pkg"

	"github.com/sirupsen/logrus"
)

func main() {
	var (
		configPath string
		h          bool
	)
	flag.StringVar(&configPath, "config", "./analyze.yaml", "config file path")
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
