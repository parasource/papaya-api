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
	"context"
	"fmt"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	defaultDb       = "dutchman"
	usersCollection = "users"
	uri             = "mongodb://localhost:27017/?maxPoolSize=20&w=majority"
)

type Config struct {
}

type Database struct {
	cfg Config

	client *mongo.Client
}

func NewDatabase(cfg Config) (*Database, error) {
	d := &Database{
		cfg: cfg,
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}
	d.client = client

	d.fillWithWardrobe()
	d.fillWithSelections()

	return d, nil
}

func (d *Database) fillWithWardrobe() {
	err := d.ClearCollection("wardrobe")
	if err != nil {
		logrus.Errorf("error clearing collection: %v", err)
	}
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
		d.StoreModel("wardrobe", interest)
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
		d.StoreModel("wardrobe", interest)
	}
}

func (d *Database) fillWithSelections() {
	err := d.ClearCollection("selections")
	if err != nil {
		logrus.Errorf("error clearing collection: %v", err)
	}

	for i := 0; i < 15; i++ {
		selection := models.FakeSelection()
		d.StoreModel("selections", selection)
	}
}

func (d *Database) GetAllSelections() ([]*models.Selection, error) {
	coll := d.client.Database(defaultDb).Collection("selections")

	var items []*models.Selection
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &items); err != nil {
		logrus.Errorf("error decoding interests: %v", err)
	}

	return items, nil
}

func (d *Database) Shutdown() {
	if err := d.client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func (d *Database) StoreModel(collection string, model interface{}) error {
	coll := d.client.Database(defaultDb).Collection(collection)

	res, err := coll.InsertOne(context.TODO(), model)
	if err != nil {
		return err
	}
	logrus.Infof("inserted doc with _id: %v", res.InsertedID)

	return nil
}

func (d *Database) ClearCollection(collection string) error {
	coll := d.client.Database(defaultDb).Collection(collection)

	_, err := coll.DeleteMany(
		context.Background(),
		bson.M{},
	)

	return err
}

// Wardrobe

func (d *Database) GetWardrobeItems() []*models.WardrobeItem {
	coll := d.client.Database(defaultDb).Collection("wardrobe")

	var interests []*models.WardrobeItem
	cursor, err := coll.Find(context.TODO(), bson.M{})
	if err != nil {
		logrus.Errorf("error cursoring all interests: %v", err)
		return nil
	}

	if err = cursor.All(context.TODO(), &interests); err != nil {
		logrus.Errorf("error decoding interests: %v", err)
	}

	return interests
}
