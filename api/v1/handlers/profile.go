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

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/parasource/papaya-api/api/v1/requests"
	database "github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/parasource/papaya-api/pkg/util"
	"github.com/sirupsen/logrus"
)

func HandleProfileSetWardrobe(c *gin.Context) {
	var r requests.SetWardrobeRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile wardrobe request: %v", err)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}
	var items []*models.WardrobeItem
	for _, itemID := range r.Wardrobe {
		items = append(items, &models.WardrobeItem{ID: itemID})
	}
	err = database.DB().Model(&user).Association("Wardrobe").Replace(items)
	if err != nil {
		logrus.Errorf("error replacing wardrobe items: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleProfileSetMood(c *gin.Context) {
	var r requests.SetMoodRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile mood request: %v", err)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	user.Mood = r.Mood
	database.DB().Save(user)

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleProfileUpdateSettings(c *gin.Context) {
	var r requests.UpdateSettingsRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding update profile settings request: %v", err)
		return
	}

	_, err = GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	//err = p.database.UpdateUserSettings(userId, &models.UserSettings{
	//	ReceivePushNotifications:  r.ReceivePushNotifications,
	//	ReceiveEmailNotifications: r.ReceiveEmailNotifications,
	//})
	user, err := GetUser(c)
	if err != nil {
		c.AbortWithStatus(403)
		return
	}

	if r.Name != "" && !util.IsLetter(r.Name) {
		c.JSON(400, gin.H{
			"success": "false",
			"message": "Недопустимый формат имени",
		})
		return
	}

	if r.Sex != "" && (r.Sex != "male" && r.Sex != "female") {
		c.AbortWithStatus(400)
		return
	}

	user.Sex = r.Sex
	if r.Name != "" {
		user.Name = r.Name
	}
	user.PushNotifications = r.ReceivePushNotifications

	err = database.DB().Save(user).Error
	if err != nil {
		logrus.Errorf("error updating user settings: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleSetAPNSToken(c *gin.Context) {
	var r requests.SetAPNSTokenRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding update profile settings request: %v", err)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	user.ApnsToken = r.ApnsToken

	err = database.DB().Save(user).Error
	if err != nil {
		logrus.Errorf("error updating user settings: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleProfileGetWardrobe(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var items []models.WardrobeItem
	err = database.DB().Raw("select * from wardrobe_items join users_wardrobe uw on wardrobe_items.id = uw.wardrobe_item_id where (wardrobe_items.sex = ? or wardrobe_items.sex = 'unisex') and uw.user_id = ?", user.Sex, user.ID).Find(&items).Error
	if err != nil {
		logrus.Errorf("error getting user's wardrobe: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, items)
}
