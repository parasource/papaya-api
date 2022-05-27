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

package papaya

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/papaya-api/api/v1/router"
	"github.com/lightswitch/papaya-api/pkg/adviser"
	"github.com/lightswitch/papaya-api/pkg/database"
	models2 "github.com/lightswitch/papaya-api/pkg/database/models"
	"github.com/sirupsen/logrus"
	"net"
)

func init() {
	gofakeit.Seed(0)
}

type Config struct {
	HttpHost        string `json:"http_host"`
	HttpPort        string `json:"http_port"`
	AdviserHost     string `json:"adviser_host"`
	AdviserPort     string `json:"adviser_port"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

type Papaya struct {
	cfg Config

	r       *gin.Engine
	adviser *adviser.Adviser
	jobs    *JobsManager
}

func NewPapaya(cfg Config, dbCfg database.Config) (*Papaya, error) {
	d := &Papaya{
		cfg: cfg,
	}

	r := router.Initialize()
	d.r = r

	// Database
	err := database.New(dbCfg)
	if err != nil {
		logrus.Fatalf("error creating database: %v", err)
	}

	adviserUrl := net.JoinHostPort(cfg.AdviserHost, cfg.AdviserPort)
	adviser.New(adviserUrl, 3)

	jobs := []*Job{
		{
			Name: "Renew today's look",
			F: func() {
				var users []models2.User
				database.DB().Find(&users)

				for _, user := range users {
					var look models2.Look

					err := database.DB().Limit(1).Order("random()").Find(&look).Error
					if err != nil {
						logrus.Errorf("error running job: %v", err)
					}

					err = database.DB().Model(&user).Association("TodayLook").Replace(&look)
					if err != nil {
						logrus.Errorf("error running job: %v", err)
					}
				}
			},
			Interval: IntervalDaily,
		},
	}
	jm, err := NewJobsManager(jobs)
	if err != nil {
		logrus.Fatalf("error running jobs manager: %v", err)
	}
	d.jobs = jm

	return d, nil
}

func (p *Papaya) Start() error {
	err := p.r.Run(net.JoinHostPort(p.cfg.HttpHost, p.cfg.HttpPort))
	if err != nil {
		logrus.Fatalf("error running gin: %v", err)
	}

	return nil
}
