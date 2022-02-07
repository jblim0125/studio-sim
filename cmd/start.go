package cmd

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jblim0125/cache-sim/common"
	"github.com/jblim0125/cache-sim/common/appdata"
	appInit "github.com/jblim0125/cache-sim/init"
)

func init() {
	startCmd.PersistentFlags().StringP("log", "l", "", "application log level(debug, info, warn, error)")
	startCmd.PersistentFlags().String("config", "./configs/dev.yaml", "path of configuration file")
	startCmd.PersistentFlags().String("dsl", "./dsl_sample/dsl_sample.json", "path of dsl file")
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Run Cache Server Client Simulator",
	RunE: func(cmd *cobra.Command, args []string) error {
		logLevel, err := cmd.Flags().GetString("log")
		if err != nil {
			return errors.Wrap(err, "can't get log level from input")
		}
		log := initLog()
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return errors.Wrap(err, "can't get path of config file")
		}
		config, err := readConfig(configPath)
		if err != nil {
			return err
		}
		setLogLevel(logLevel, &config.Log)
		appContext := appInit.Context{}.New(log, config)
		dslPath, err := cmd.Flags().GetString("dsl")
		if err != nil {
			return errors.Wrap(err, "can't get path of config file")
		}
		if err := appContext.Initialize(dslPath); err != nil {
			return err
		}
		appContext.StartSubModules()
		log.Shutdown()
		return nil
	},
}

func initLog() *common.Logger {
	log := common.Logger{}.GetInstance()
	log.Start()
	return log
}

func readConfig(path string) (*appdata.Configuration, error) {
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

func setLogLevel(input string, conf *appdata.LogConfiguration) {
	log := common.Logger{}.GetInstance()
	if len(input) > 0 {
		_, err := logrus.ParseLevel(input)
		if err == nil {
			conf.Level = input
		}
	}
	log.Setting(conf)
}
