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
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (d *Database) StoreUser(user *models.User) error {
	return d.StoreModel(usersCollection, user)
}

func (d *Database) CheckIfUserExists(email string) bool {
	coll := d.client.Database(defaultDb).Collection(usersCollection)

	count, err := coll.CountDocuments(context.TODO(), bson.M{"email": email})
	if err != nil {
		logrus.Errorf("error counting users: %v", err)
		return true
	}

	return count > 0
}

func (d *Database) GetUser(userId string) *models.User {
	coll := d.client.Database(defaultDb).Collection(usersCollection)

	res := coll.FindOne(context.TODO(), bson.M{"_id": userId})
	if res.Err() != nil {
		logrus.Errorf("error getting result from mongo: %v", res.Err())
		return nil
	}

	var user models.User
	err := res.Decode(&user)
	if err != nil {
		logrus.Errorf("error decoding result: %v", err)
		return nil
	}

	return &user
}

func (d *Database) GetUserByEmail(email string) *models.User {
	coll := d.client.Database(defaultDb).Collection(usersCollection)

	res := coll.FindOne(context.TODO(), bson.M{"email": email})
	if res.Err() != nil {
		logrus.Errorf("error getting result from mongo: %v", res.Err())
		return nil
	}

	var user models.User
	err := res.Decode(&user)
	if err != nil {
		logrus.Errorf("error decoding result: %v", err)
		return nil
	}

	return &user
}

func (d *Database) SetUserMood(userId string, mood string) error {
	coll := d.client.Database(defaultDb).Collection(usersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := coll.UpdateOne(
		ctx,
		bson.M{
			"_id": userId,
		},
		bson.D{
			{
				"$set", bson.D{{"mood", mood}},
			},
		},
	)

	return err
}

func (d *Database) SetUserWardrobe(userId string, items []string) error {
	coll := d.client.Database(defaultDb).Collection(usersCollection)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := coll.UpdateOne(
		ctx,
		bson.M{
			"_id": userId,
		},
		bson.D{
			{"$set", bson.D{{"wardrobe", items}}},
		},
	)

	return err
}
