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
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/papaya/models"
	"github.com/lightswitch/dutchman-backend/papaya/requests"
	"github.com/sirupsen/logrus"
)

func (p *Papaya) HandleProfileSetWardrobe(c *gin.Context) {
	var r requests.SetWardrobeRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile wardrobe request: %v", err)
		return
	}

	user, err := p.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}
	var items []*models.WardrobeItem
	for _, itemID := range r.Wardrobe {
		items = append(items, &models.WardrobeItem{ID: itemID})
	}
	err = p.db.DB().Model(&user).Association("Wardrobe").Replace(items)
	if err != nil {
		logrus.Errorf("error replacing wardrobe items: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (p *Papaya) HandleProfileSetMood(c *gin.Context) {
	var r requests.SetMoodRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile mood request: %v", err)
		return
	}

	user, err := p.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	user.Mood = r.Mood
	p.db.DB().Save(user)

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (p *Papaya) HandleProfileUpdateSettings(c *gin.Context) {
	var r requests.UpdateSettingsRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding update profile settings request: %v", err)
		return
	}

	_, err = p.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	//
	//err = p.db.UpdateUserSettings(userId, &models.UserSettings{
	//	ReceivePushNotifications:  r.ReceivePushNotifications,
	//	ReceiveEmailNotifications: r.ReceiveEmailNotifications,
	//})
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (p *Papaya) HandleProfileGetWardrobe(c *gin.Context) {
	user, err := p.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	c.JSON(200, user.Wardrobe)
}
