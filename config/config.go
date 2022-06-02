package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/ssrlive/proxypool/log"

	"github.com/ghodss/yaml"
	"github.com/ssrlive/proxypool/pkg/tool"
)

var configFilePath = "config.yaml"

type ConfigOptions struct {
	Domain                string   `json:"domain" yaml:"domain"`
	Port                  string   `json:"port" yaml:"port"`
	DatabaseUrl           string   `json:"database_url" yaml:"database_url"`
	CrawlInterval         uint64   `json:"crawl-interval" yaml:"crawl-interval"`
	CFEmail               string   `json:"cf_email" yaml:"cf_email"`
	CFKey                 string   `json:"cf_key" yaml:"cf_key"`
	SourceFiles           []string `json:"source-files" yaml:"source-files"`
	HealthCheckTimeout    int      `json:"healthcheck-timeout" yaml:"healthcheck-timeout"`
	HealthCheckConnection int      `json:"healthcheck-connection" yaml:"healthcheck-connection"`
	SpeedTest             bool     `json:"speedtest" yaml:"speedtest"`
	SpeedTestInterval     uint64   `json:"speedtest-interval" yaml:"speedtest-interval"`
	SpeedTimeout          int      `json:"speed-timeout" yaml:"speed-timeout"`
	SpeedConnection       int      `json:"speed-connection" yaml:"speed-connection"`
	ActiveFrequency       uint16   `json:"active-frequency" yaml:"active-frequency" `
	ActiveInterval        uint64   `json:"active-interval" yaml:"active-interval"`
	ActiveMaxNumber       uint16   `json:"active-max-number" yaml:"active-max-number"`
}

// Config 配置
var Config ConfigOptions

func (config ConfigOptions) HostUrl() string {
	url := config.Domain
	if len(strings.Split(url, ":")) <= 1 {
		url = url + ":" + config.Port
	}
	return url
}

// Parse 解析配置文件，支持本地文件系统和网络链接
func Parse(path string) error {
	if path == "" {
		path = configFilePath
	} else {
		configFilePath = path
	}
	fileData, err := ReadFile(path)
	if err != nil {
		return err
	}
	Config = ConfigOptions{}
	err = yaml.Unmarshal(fileData, &Config)
	if err != nil {
		return err
	}

	// set default
	if Config.SpeedConnection <= 0 {
		Config.SpeedConnection = 5
	}
	// set default
	if Config.HealthCheckConnection <= 0 {
		Config.HealthCheckConnection = 500
	}
	if Config.Port == "" {
		Config.Port = "12580"
	}
	if Config.CrawlInterval == 0 {
		Config.CrawlInterval = 60
	}
	if Config.SpeedTestInterval == 0 {
		Config.SpeedTestInterval = 720
	}
	if Config.ActiveInterval == 0 {
		Config.ActiveInterval = 60
	}
	if Config.ActiveFrequency == 0 {
		Config.ActiveFrequency = 100
	}
	if Config.ActiveMaxNumber == 0 {
		Config.ActiveMaxNumber = 100
	}

	// 部分配置环境变量优先
	if domain := os.Getenv("DOMAIN"); domain != "" {
		Config.Domain = domain
	}
	if cfEmail := os.Getenv("CF_API_EMAIL"); cfEmail != "" {
		Config.CFEmail = cfEmail
	}
	if cfKey := os.Getenv("CF_API_KEY"); cfKey != "" {
		Config.CFKey = cfKey
	}
	s, _ := json.Marshal(Config)
	log.Debugln("Config options: %s", string(s))

	return nil
}

func IsLocalFile(path string) bool {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return false
	}
	return true
}

// 从本地文件或者http链接读取配置文件内容
func ReadFile(path string) ([]byte, error) {
	if !IsLocalFile(path) {
		resp, err := tool.GetHttpClient().Get(path)
		if err != nil {
			return nil, errors.New("config file http get fail")
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	} else {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, err
		}
		return ioutil.ReadFile(path)
	}
}

func fullDirOfExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func moduleName() string {
	info, _ := debug.ReadBuildInfo()
	_, name := filepath.Split(info.Main.Path)
	return name
}

// 返回資源文件所在的根目錄.
func ResourceRoot() string {
	exe, _ := os.Executable()
	_, file := filepath.Split(exe)

	currDir, _ := os.Getwd()
	exeDir := fullDirOfExecutable()
	if exeDir != currDir {
		// 從 go run 運行, 或者從 別的目錄 運行.
		module := moduleName()
		os := runtime.GOOS
		if os == "windows" {
			module = module + ".exe"
		}
		if file == module {
			// 可執行文件在別的目錄運行.
			return exeDir
		} else {
			// 從 go run 運行, 可執行文件生成在臨時目錄,
			// 於是返回當前目錄作爲資源根目錄.
			return currDir
		}
	} else {
		// 從 exe 所在目錄運行.
		return exeDir
	}
}
