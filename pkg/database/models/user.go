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

package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name      string `json:"name"`
	Email     string `json:"email" gorm:"unique"`
	Password  string `json:"-"`
	ApnsToken string `json:"apns_token"`
	FcmToken  string `json:"fcm_token"`

	Sex    string `json:"sex"`
	Age    int    `json:"age"`
	Avatar string `json:"avatar"`

	Wardrobe []*WardrobeItem `json:"wardrobe" gorm:"many2many:users_wardrobe;"`
	Mood     string          `json:"mood"`

	SavedTopics   []*Topic `json:"saved_topics" gorm:"many2many:saved_topics;"`
	LikedLooks    []*Look  `json:"-" gorm:"many2many:liked_looks;"`
	DislikedLooks []*Look  `json:"-" gorm:"many2many:disliked_looks;"`
	TodayLooks    []*Look  `json:"today_look" gorm:"many2many:today_looks;"`
	SavedLooks    []*Look  `json:"-" gorm:"many2many:saved_looks;"`

	EmailNotifications bool `json:"email_notifications"`
	PushNotifications  bool `json:"push_notifications"`
}

func NewUser(email string, name string, password string) *User {
	pwd, _ := hashPassword(password)
	return &User{
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
