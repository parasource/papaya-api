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

package database

import (
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	defaultDb       = "dutchman"
	usersCollection = "users"
)

type Config struct {
}

type Database struct {
	cfg Config

	db *gorm.DB
}

func NewDatabase(cfg Config) (*Database, error) {
	d := &Database{
		cfg: cfg,
	}

	dsn := "host=localhost user=dutchman password=admin dbname=dutchman port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Fatalf("error connecting to postgres: %v", err)
	}
	d.db = db

	err = d.setup()
	if err != nil {
		logrus.Fatalf("error migrating: %v", err)
	}

	return d, nil
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func (d *Database) GetUserByEmail(email string) *models.User {
	var user models.User

	d.db.Preload("Wardrobe").Preload("FavoriteLooks").Preload("FavoriteSelections").First(&user, "email = ?", email)
	if user.ID == 0 {
		return nil
	}

	return &user
}

func (d *Database) GetUser(id uint) *models.User {
	var user models.User

	d.db.First(&user, "id = ?", id)
	if user.ID == 0 {
		return nil
	}

	return &user
}

func (d *Database) CreateUser(user *models.User) {
	d.db.Create(user)
}

func (d *Database) setup() error {

	err := d.db.AutoMigrate(
		&models.User{},
		&models.WardrobeCategory{},
		&models.WardrobeItem{},
		&models.Selection{},
		&models.Look{},
		&models.LookItem{},
	)
	if err != nil {
		return err
	}

	var count int64
	if err := d.db.Model(&models.WardrobeItem{}).Count(&count).Error; err == nil && count == 0 {
		d.seed()
	}

	return err
}

func (d *Database) seed() {
	//////////////////////
	// WARDROBE

	cat1 := models.WardrobeCategory{
		Name: "Верхняя одежда",
		Slug: "cloths",
	}
	d.db.Create(&cat1)
	cat2 := models.WardrobeCategory{
		Name: "Штаны",
		Slug: "trousers",
	}
	d.db.Create(&cat2)
	cat3 := models.WardrobeCategory{
		Name: "Обувь",
		Slug: "shoes",
	}
	d.db.Create(&cat3)

	d.db.Create(&models.WardrobeItem{
		Name:               "Белая майка",
		Slug:               "white-tshirt",
		WardrobeCategoryID: cat1.ID,
	})
	d.db.Create(&models.WardrobeItem{
		Name:               "Черная майка",
		Slug:               "black-tshirt",
		WardrobeCategoryID: cat1.ID,
	})
	d.db.Create(&models.WardrobeItem{
		Name:               "Брюки",
		Slug:               "trousers",
		WardrobeCategoryID: cat2.ID,
	})
	d.db.Create(&models.WardrobeItem{
		Name:               "Джинсы",
		Slug:               "jeans",
		WardrobeCategoryID: cat2.ID,
	})
	d.db.Create(&models.WardrobeItem{
		Name:               "Кроссовки",
		Slug:               "jeans",
		WardrobeCategoryID: cat3.ID,
	})
	d.db.Create(&models.WardrobeItem{
		Name:               "Джинсы",
		Slug:               "jeans",
		WardrobeCategoryID: cat3.ID,
	})

	//////////////////////
	// LOOKS

	look1 := &models.Look{
		Name:  "Первый лук",
		Slug:  "test",
		Image: "test.jpg",
		Desc:  "Первый лук",
	}
	d.db.Create(look1)
	look2 := &models.Look{
		Name:  "Второй лук",
		Slug:  "test1",
		Image: "test.jpg",
		Desc:  "Второй лук",
	}
	d.db.Create(look2)
	look3 := &models.Look{
		Name:  "Третий лук",
		Slug:  "test2",
		Image: "test.jpg",
		Desc:  "Третий лук",
	}
	d.db.Create(look3)

	d.db.Create(&models.Selection{
		Name:  "Летняя подборка",
		Slug:  "summer",
		Desc:  "Летняя подборка",
		Looks: []*models.Look{look1, look3},
	})
	d.db.Create(&models.Selection{
		Name:  "Зимняя подборка",
		Slug:  "winter",
		Desc:  "Зимняя подборка",
		Looks: []*models.Look{look1, look2, look3},
	})
}
