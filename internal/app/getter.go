package app

import (
	"errors"
	"github.com/asdlokj1qpi23/proxypool/config"
	"github.com/asdlokj1qpi23/proxypool/internal/cache"
	"github.com/asdlokj1qpi23/proxypool/log"
	"github.com/asdlokj1qpi23/proxypool/pkg/getter"
	"github.com/ghodss/yaml"
	"path/filepath"
)

var Getters = make([]getter.Getter, 0)

func InitConfigAndGetters() (err error) {
	err = config.Parse()
	if err != nil {
		return
	}
	if s := config.Config.SourceFiles; len(s) == 0 {
		return errors.New("no sources")
	} else {
		for index, path := range s {
			if config.IsLocalFile(path) && !filepath.IsAbs(path) {
				var configDir = filepath.Dir(config.FilePath())
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

func CheckPoolAvailable() {
	err := config.Parse()
	if err != nil {
		panic("config err")
	}
	if config.Config.PoolFilesCheck {
		log.Infoln("pool file check start ")
		if s := config.Config.PoolFiles; len(s) != 0 {
			for index, pool := range s {
				if config.IsLocalFile(pool.Url) && !filepath.IsAbs(pool.Url) {
					var configDir = filepath.Dir(config.FilePath())
					s[index] = config.PoolFile{Type: pool.Type, Url: filepath.Join(configDir, pool.Url)}
				}
			}
			checkPoolFileList(s)
		}
		log.Infoln("pool file check end ")
	}
}

func checkPoolFileList(files []config.PoolFile) {
	for _, path := range files {
		data, err := config.ReadFile(path.Url)
		if err != nil {
			log.Errorln("Init SourceFile Error: %s\n", err.Error())
			continue
		}
		var configDir = filepath.Dir(config.FilePath())
		log.Infoln(configDir, string(data))
		//err = os.WriteFile("output.txt", data, 0644)
		//if err != nil {
		//	panic(err)
		//}
		//sourceList := make([]config.Source, 0)
		//err = yaml.Unmarshal(data, &sourceList)
		//if err != nil {
		//	log.Errorln("Init SourceFile Error: %s\n", err.Error())
		//	continue
		//}
		//for _, source := range sourceList {
		//	if source.Options == nil {
		//		continue
		//	}
		//	g, err := getter.NewGetter(source.Type, source.Options)
		//	if err == nil && g != nil {
		//		Getters = append(Getters, g)
		//		log.Debugln("init getter: %s %v", source.Type, source.Options)
		//	}
		//}
	}
}
