package pkg

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"prom-metric-analyze/config"

	"github.com/sirupsen/logrus"
)

const (
	downloadURL = "https://github.com/grafana/mimir/releases/download/mimir-2.6.0/mimirtool-linux-amd64"
)

func CheckMimirToolBinary() (*string, error) {
	// 判断是否已经存在二进制
	p := config.Get().MimirtoolDIR
	binaryPath := fmt.Sprintf("%s/mimirtool", p)
	_, err := os.Stat(binaryPath)
	if err == nil || os.IsExist(err) {
		logrus.Warnln(binaryPath + " mimirtool binary is exist")
		return &binaryPath, nil
	}

	os.Mkdir(p, 0755)

	logrus.Warnln("begin to download mimirtool", binaryPath, downloadURL)
	resp, err := http.Get(downloadURL)
	if err != nil {
		logrus.Errorln("download mimirtool failed", err)
		return nil, err
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorln("io.ReadAll mimirtool failed", err)
		return nil, err
	}

	if err := os.WriteFile(binaryPath, bs, 0755); err != nil {
		logrus.Errorln("os.WriteFile mimirtool failed", err)
		return nil, err
	}
	return &binaryPath, nil
}
