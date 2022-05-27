/*
 * Copyright 2022 Parasource Organization
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/lightswitch/papaya-api/pkg"
	"github.com/lightswitch/papaya-api/pkg/database"
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

	"db_host":     "database",
	"db_port":     "5432",
	"db_database": "papaya",
	"db_user":     "papaya-api",
	"db_password": "6TKE+J({K$>*^fD",

	"adviser_host": "gorse-server",
	"adviser_port": "8087",

	// seconds to wait til force shutdown
	"shutdown_timeout": 30,
}

func init() {
	rootCmd.Flags().String("http_host", "127.0.0.1", "file server http host")
	rootCmd.Flags().String("http_port", "8000", "file server http port")
	rootCmd.Flags().String("db_host", "localhost", "database host")
	rootCmd.Flags().String("db_port", "5432", "database port")
	rootCmd.Flags().String("db_database", "papaya", "database name")
	rootCmd.Flags().String("db_user", "", "database username")
	rootCmd.Flags().String("db_password", "", "file server http port")
	rootCmd.Flags().String("adviser_host", "gorse-server", "adviser host")
	rootCmd.Flags().String("adviser_port", "8087", "adviser port")
	rootCmd.Flags().Int("shutdown_timeout", 30, "node graceful shutdown timeout")

	viper.BindPFlag("http_host", rootCmd.Flags().Lookup("http_host"))
	viper.BindPFlag("http_port", rootCmd.Flags().Lookup("http_port"))
	viper.BindPFlag("db_host", rootCmd.Flags().Lookup("db_host"))
	viper.BindPFlag("db_port", rootCmd.Flags().Lookup("db_port"))
	viper.BindPFlag("db_database", rootCmd.Flags().Lookup("db_database"))
	viper.BindPFlag("db_user", rootCmd.Flags().Lookup("db_user"))
	viper.BindPFlag("db_password", rootCmd.Flags().Lookup("db_password"))
	viper.BindPFlag("adviser_host", rootCmd.Flags().Lookup("adviser_host"))
	viper.BindPFlag("adviser_port", rootCmd.Flags().Lookup("adviser_port"))
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
			"db_host", "db_port", "db_database", "db_user", "db_password",
			"adviser_host", "adviser_port",
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

		dbHost := v.GetString("db_host")
		dbPort := v.GetString("db_port")
		dbDatabase := v.GetString("db_database")
		dbUser := v.GetString("db_user")
		dbPass := v.GetString("db_password")

		adviserHost := v.GetString("adviser_host")
		adviserPort := v.GetString("adviser_port")

		papaya, err := papaya.NewPapaya(papaya.Config{
			HttpHost:    httpHost,
			HttpPort:    httpPort,
			AdviserHost: adviserHost,
			AdviserPort: adviserPort,
		}, database.Config{
			Host:     dbHost,
			Port:     dbPort,
			Database: dbDatabase,
			User:     dbUser,
			Password: dbPass,
		})
		if err != nil {
			logrus.Fatal(err)
		}
		err = papaya.Start()
		if err != nil {
			logrus.Fatalf("error starting papaya: %v", err)
		}

	},
}

func printWelcome() {
	text := "    ____                               \n   / __ \\____ _____  ____ ___  ______ _\n  / /_/ / __ `/ __ \\/ __ `/ / / / __ `/\n / ____/ /_/ / /_/ / /_/ / /_/ / /_/ / \n/_/    \\__,_/ .___/\\__,_/\\__, /\\__,_/  \n           /_/          /____/         "

	println(text)
}
