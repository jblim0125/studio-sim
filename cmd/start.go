package cmd

import (
	appInit "github.com/jblim0125/cache-sim/init"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	startCmd.PersistentFlags().StringP("log", "l", "", "application log level(debug, info, warn, error)")
	startCmd.PersistentFlags().String("config", "configs/dev.yaml", "path of configuration file")
	startCmd.PersistentFlags().String("dsl", "dsl_sample/dsl_sample.json", "path of dsl file")
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Run Cache Server Client Simulator",
	RunE: func(cmd *cobra.Command, args []string) error {

		appContext := appInit.Context{}.New()
		appContext.InitLog()
		logLevel, err := cmd.Flags().GetString("log")
		if err != nil {
			return errors.Wrap(err, "can't get log level from input")
		}
		appContext.SetLogLevel(logLevel)

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			return errors.Wrap(err, "can't get path of config file")
		}
		appContext.ReadConfig(configPath)

		dslPath, err := cmd.Flags().GetString("dsl")
		if err != nil {
			return errors.Wrap(err, "can't get path of config file")
		}

		appContext.Initialize(dslPath)

		appContext.StartSubModules()

		appContext.Log.Shutdown()
		return nil
	},
}
