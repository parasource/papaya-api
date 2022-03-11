/*
 * Copyright 2022 LightSwitch.Digital
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

package dutchman

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/dutchman/database"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/sirupsen/logrus"
	"net"
)

type Config struct {
	HttpHost        string `json:"http_host"`
	HttpPort        string `json:"http_port"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

type Dutchman struct {
	cfg Config

	r  *gin.Engine
	db *database.Database
}

func NewDutchman(cfg Config) (*Dutchman, error) {
	d := &Dutchman{
		cfg: cfg,
	}

	// Gin server
	r := gin.Default()
	d.registerRoutes(r)
	d.r = r

	// Database
	db, err := database.NewDatabase(database.Config{})
	if err != nil {
		logrus.Fatalf("error creating database: %v", err)
	}
	d.db = db

	for i := 0; i < 15; i++ {
		interest := &models.WardrobeItem{
			ID:       fmt.Sprintf("shs%v", i),
			Name:     fmt.Sprintf("Ботинки %v", i),
			Slug:     fmt.Sprintf("shoes-%v", i),
			Category: "shoes",
			Sex: []string{
				"male",
			},
		}
		d.db.StoreModel("wardrobe", interest)
	}
	for i := 0; i < 15; i++ {
		interest := &models.WardrobeItem{
			ID:       fmt.Sprintf("hts-%v", i),
			Name:     fmt.Sprintf("Шапки %v", i),
			Slug:     fmt.Sprintf("hats-%v", i),
			Category: "hats",
			Sex: []string{
				"female",
			},
		}
		d.db.StoreModel("wardrobe", interest)
	}

	return d, nil
}

func (d *Dutchman) Start() error {
	err := d.r.Run(net.JoinHostPort(d.cfg.HttpHost, d.cfg.HttpPort))
	if err != nil {
		logrus.Fatalf("error running gin: %v", err)
	}

	return nil
}
