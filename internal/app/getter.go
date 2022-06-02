package app

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ssrlive/proxypool/log"

	"github.com/ssrlive/proxypool/internal/cache"

	"github.com/ghodss/yaml"

	"github.com/ssrlive/proxypool/config"
	"github.com/ssrlive/proxypool/pkg/getter"
)

var Getters = make([]getter.Getter, 0)

var configFilePath = ""

func SetConfigFilePath(path string) {
	configFilePath = path
}

func ConfigFilePath() string {
	return configFilePath
}

func configFileFullPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	exPath, _ := os.Getwd()

	return filepath.Join(exPath, path)
}

func InitConfigAndGetters(path string) (err error) {
	var configDir string
	if config.IsLocalFile(path) {
		path = configFileFullPath(path)
		configDir = filepath.Dir(path)
	}
	err = config.Parse(path)
	if err != nil {
		return
	}
	if s := config.Config.SourceFiles; len(s) == 0 {
		return errors.New("no sources")
	} else {
		for index, path := range s {
			if config.IsLocalFile(path) && !filepath.IsAbs(path) {
				s[index] = filepath.Join(configDir, path)
			}
		}
		initGetters(s)
	}
	return
}

func initGetters(sourceFiles []string) {
	Getters = make([]getter.Getter, 0)
	for _, path := range sourceFiles {
		data, err := config.ReadFile(path)
		if err != nil {
			log.Errorln("Init SourceFile Error: %s\n", err.Error())
			continue
		}
		sourceList := make([]config.Source, 0)
		err = yaml.Unmarshal(data, &sourceList)
		if err != nil {
			log.Errorln("Init SourceFile Error: %s\n", err.Error())
			continue
		}
		for _, source := range sourceList {
			if source.Options == nil {
				continue
			}
			g, err := getter.NewGetter(source.Type, source.Options)
			if err == nil && g != nil {
				Getters = append(Getters, g)
				log.Debugln("init getter: %s %v", source.Type, source.Options)
			}
		}
	}
	log.Infoln("Getter count: %d", len(Getters))
	cache.GettersCount = len(Getters)
}
