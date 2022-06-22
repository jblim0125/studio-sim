package cmd

import (
	"path/filepath"
	"strings"

	"github.com/jblim0125/cache-sim/common"
	"github.com/jblim0125/cache-sim/common/appdata"
)

func InitLog() *common.Logger {
	log := common.Logger{}.GetInstance()
	log.Start()
	return log
}

func ReadConfig(path string) (*appdata.Configuration, error) {
	log := common.Logger{}.GetInstance()
	configManager := common.ConfigManager{}.New(log)
	// Path
	configPath := filepath.Dir(path)
	configName := strings.TrimLeft(path, configPath)
	configType := "yaml"
	// Config file struct
	conf := new(appdata.Configuration)
	// Read
	if err := configManager.ReadConfig(configPath, configName, configType, conf); err != nil {
		return nil, err
	}
	log.Error("Running Option")
	conf.Print(log.Out)
	log.Errorf("[ Configuration ] Read ............................................................ [ OK ]")
	return conf, nil
}

// SetLogLevel set log level
func SetLogLevel(input string) {
	log := common.Logger{}.GetInstance()
	log.SetLogLevelByString(input)
}
