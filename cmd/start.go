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
	startCmd.PersistentFlags().StringP("log", "l", "debug", "application log level(debug, info, warn, error)")
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
		log, err := initLog(logLevel)
		if err != nil {
			return err
		}
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return errors.Wrap(err, "can't get path of config file")
		}
		config, err := readConfig(configPath)
		if err != nil {
			return err
		}
		appContext := appInit.Context{}.New(log, config)
		if err := appContext.SetLogger(); err != nil {
			return err
		}
		dslPath, err := cmd.Flags().GetString("dsl")
		if err != nil {
			return errors.Wrap(err, "can't get path of config file")
		}
		if err := appContext.Initialize(dslPath); err != nil {
			return err
		}
		return nil
	},
}

func initLog(level string) (*common.Logger, error) {
	log := common.Logger{}.GetInstance()
	lv, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, errors.Wrap(err, "init log(get log level)")
	}
	log.SetLogLevel(lv)
	log.Start()
	return log, nil
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
