package main

import (
	"github.com/lightswitch/dutchman-backend/dutchman"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		dutchman, err := dutchman.NewDutchman(dutchman.Config{
			HttpHost: "localhost",
			HttpPort: "8000",
		})
		if err != nil {
			panic(err)
		}
		err = dutchman.Start()
		if err != nil {
			logrus.Fatalf("error starting dutchman: %v", err)
		}

		println("success")
	},
}
