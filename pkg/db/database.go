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

package database

import (
	"fmt"
	models2 "github.com/lightswitch/papaya-api/pkg/db/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var conn *gorm.DB

type Config struct {
	Host     string
	Port     string
	Database string
	User     string
	Password string
}

type Database struct {
	cfg Config

	db *gorm.DB
}

func New(cfg Config) error {
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port)

	var (
		db      *gorm.DB
		err     error
		retries = 0
	)
	for {
		if retries >= 3 {
			logrus.Fatalf("error connecting to postgres")
		}

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}

		retries++
		<-time.After(time.Second)
	}
	if err != nil {
		logrus.Fatalf("error connecting to postgres: %v", err)
	}

	err = migrate(db)
	if err != nil {
		logrus.Fatalf("error migrating: %v", err)
	}

	conn = db

	return err
}

func DB() *gorm.DB {
	return conn
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func GetUserByEmail(email string) *models2.User {
	var user models2.User

	conn.Preload("Wardrobe").Preload("SavedTopics").Preload("Collections").First(&user, "email = ?", email)
	if user.ID == 0 {
		return nil
	}

	return &user
}

func GetUser(id uint) *models2.User {
	var user models2.User

	conn.First(&user, "id = ?", id)
	if user.ID == 0 {
		return nil
	}

	return &user
}

func CreateUser(user *models2.User) {
	conn.Create(user)
}

func migrate(db *gorm.DB) error {

	err := db.AutoMigrate(
		&models2.User{},
		&models2.WardrobeCategory{},
		&models2.WardrobeItem{},
		&models2.Tag{},
		&models2.Look{},
		&models2.ItemURL{},
		&models2.Category{},
		&models2.Collection{},
		&models2.SearchRecord{},
	)
	if err != nil {
		return err
	}

	return err
}
