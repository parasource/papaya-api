/*
 * Copyright 2023 Parasource Organization
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
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

const (
	setupScript = `

	UPDATE looks SET tsv = to_tsvector('russian', looks.name) || to_tsvector('russian', looks.desc) WHERE tsv IS NULL;
	UPDATE topics SET tsv = to_tsvector('russian', topics.name) || to_tsvector('russian', topics.desc) WHERE tsv IS NULL;
	UPDATE wardrobe_items SET tsv = to_tsvector('russian', wardrobe_items.name) WHERE tsv IS NULL;

	CREATE INDEX IF NOT EXISTS idx_tsv_looks ON looks USING gin(tsv);
	CREATE INDEX IF NOT EXISTS idx_tsv_topics ON topics USING gin(tsv);
	CREATE INDEX IF NOT EXISTS idx_tsv_wardrobe_items ON wardrobe_items USING gin(tsv);

	CREATE INDEX IF NOT EXISTS idx_search_records ON search_records (lower(query) text_pattern_ops);

	/* ------------------- */
	/* UPDATE TSV TRIGGERS */

	DROP TRIGGER IF EXISTS looks_tsv_insert on looks;
	CREATE TRIGGER looks_tsv_insert BEFORE INSERT OR UPDATE
    ON looks
    FOR EACH ROW EXECUTE PROCEDURE
    tsvector_update_trigger(tsv, 'pg_catalog.russian', name, "desc");

	DROP TRIGGER IF EXISTS topics_tsv_insert on topics;
	CREATE TRIGGER topics_tsv_insert BEFORE INSERT OR UPDATE
    ON topics
    FOR EACH ROW EXECUTE PROCEDURE
    tsvector_update_trigger(tsv, 'pg_catalog.russian', name, "desc");

	DROP TRIGGER IF EXISTS wardrobe_items_tsv_insert on wardrobe_items;
	CREATE TRIGGER wardrobe_items_tsv_insert BEFORE INSERT OR UPDATE
    ON wardrobe_items
    FOR EACH ROW EXECUTE PROCEDURE
    tsvector_update_trigger(tsv, 'pg_catalog.russian', name, tags);


	/* ------------ */
	/* SEARCH VIEWS */

	--- For backward compatibility

	CREATE OR REPLACE VIEW searches_male AS

    SELECT text 'looks' as origin_table, looks.id, looks.tsv
	FROM looks
	WHERE looks.sex = 'male' AND looks.deleted_at IS NULL
	GROUP BY looks.id, text 'looks', looks.tsv

    UNION ALL

    SELECT text 'topics' as origin_table, id, tsv
    FROM topics;

	CREATE OR REPLACE VIEW searches_female AS

    SELECT text 'looks' as origin_table, looks.id, looks.tsv
	FROM looks
	WHERE looks.sex = 'female' AND looks.deleted_at IS NULL
	GROUP BY looks.id, text 'looks', looks.tsv

    UNION ALL

    SELECT text 'topics' as origin_table, id, tsv
    FROM topics;
	`
)

var conn *gorm.DB

type Config struct {
	Address string
}

type Database struct {
	cfg Config

	db *gorm.DB
}

func New(cfg Config) error {
	var (
		db      *gorm.DB
		err     error
		retries = 0
	)
	for {
		if retries >= 3 {
			logrus.Fatalf("error connecting to postgres")
		}

		db, err = gorm.Open(postgres.Open(cfg.Address), &gorm.Config{})
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

	// Running database setup script
	err = db.Exec(setupScript).Error
	if err != nil {
		logrus.Errorf("error running sql setup script: %v", err)
	} else {
		logrus.Infof("database setup script run successfully")
	}

	conn = db

	return nil
}

func DB() *gorm.DB {
	return conn
}

func (d *Database) DB() *gorm.DB {
	return d.db
}

func GetUserByEmail(email string) *models.User {
	var user models.User

	conn.Preload("Wardrobe").Preload("SavedTopics").First(&user, "email = ?", email)
	if user.ID == 0 {
		return nil
	}

	return &user
}

func GetUser(id uint) *models.User {
	var user models.User

	conn.First(&user, "id = ?", id)
	if user.ID == 0 {
		return nil
	}

	return &user
}

func CreateUser(user *models.User) {
	conn.Create(user)
}

func migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&models.User{},
		&models.WardrobeCategory{},
		&models.WardrobeItem{},
		&models.Look{},
		&models.Topic{},
		&models.ItemURL{},
		&models.Category{},
		&models.SearchRecord{},
		&models.Article{},
	)
	if err != nil {
		return err
	}

	return err
}
