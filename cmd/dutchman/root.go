package main

import (
	"github.com/lightswitch/dutchman-backend/dutchman"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"runtime"
)

var configDefaults = map[string]interface{}{
	"gomaxprocs": 0,
	// http file server host and port
	"http_host": "127.0.0.1",
	"http_port": "8000",

	// seconds to wait til force shutdown
	"shutdown_timeout": 30,
}

func init() {
	rootCmd.Flags().String("http_host", "127.0.0.1", "file server http host")
	rootCmd.Flags().String("http_port", "8000", "file server http port")
	rootCmd.Flags().Int("shutdown_timeout", 30, "node graceful shutdown timeout")

	viper.BindPFlag("http_host", rootCmd.Flags().Lookup("http_host"))
	viper.BindPFlag("http_port", rootCmd.Flags().Lookup("http_port"))
	viper.BindPFlag("shutdown_timeout", rootCmd.Flags().Lookup("shutdown_timeout"))
}

var rootCmd = &cobra.Command{
	Run: func(cmd *cobra.Command, args []string) {
		printWelcome()

		for k, v := range configDefaults {
			viper.SetDefault(k, v)
		}

		bindEnvs := []string{
			"http_host", "http_port",
			"shutdown_timeout",
		}
		for _, env := range bindEnvs {
			err := viper.BindEnv(env)
			if err != nil {
				logrus.Fatalf("error binding env variable: %v", err)
			}
		}

		if os.Getenv("GOMAXPROCS") == "" {
			if viper.IsSet("gomaxprocs") && viper.GetInt("gomaxprocs") > 0 {
				runtime.GOMAXPROCS(viper.GetInt("gomaxprocs"))
			} else {
				runtime.GOMAXPROCS(runtime.NumCPU())
			}
		}

		v := viper.GetViper()

		httpHost := v.GetString("http_host")
		httpPort := v.GetString("http_port")

		dutchman, err := dutchman.NewDutchman(dutchman.Config{
			HttpHost: httpHost,
			HttpPort: httpPort,
		})
		if err != nil {
			panic(err)
		}
		err = dutchman.Start()
		if err != nil {
			logrus.Fatalf("error starting dutchman: %v", err)
		}

	},
}

func printWelcome() {
	text := "    ____  __  ________________  ____  ______    _   __\n   / __ \\/ / / /_  __/ ____/ / / /  |/  /   |  / | / /\n  / / / / / / / / / / /   / /_/ / /|_/ / /| | /  |/ / \n / /_/ / /_/ / / / / /___/ __  / /  / / ___ |/ /|  /  \n/_____/\\____/ /_/  \\____/_/ /_/_/  /_/_/  |_/_/ |_/   \n                                                      "

	println(text)
}
