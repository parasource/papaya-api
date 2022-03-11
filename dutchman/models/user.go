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

package models

import (
	"github.com/lightswitch/dutchman-backend/dutchman/util/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserSettings struct {
	ReceiveNotifications bool `bson:"receive_notifications"`
}

type User struct {
	ID       string `json:"_id" bson:"_id"`
	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"-" bson:"password"`

	Wardrobe []string `json:"wardrobe" bson:"wardrobe"`
	Mood     string   `json:"mood" bson:"mood"`

	Settings UserSettings `json:"settings" bson:"settings"`
}

func NewUser(email string, name string, password string) *User {
	uid, _ := uuid.NewV4()
	pwd, _ := hashPassword(password)
	return &User{
		ID:       uid.String(),
		Name:     name,
		Email:    email,
		Password: pwd,
	}
}

func (u *User) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}
