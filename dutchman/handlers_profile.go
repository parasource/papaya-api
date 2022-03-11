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

package dutchman

import (
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/lightswitch/dutchman-backend/dutchman/requests"
	"github.com/lightswitch/dutchman-backend/dutchman/util"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (d *Dutchman) HandleProfileSetWardrobe(c *gin.Context) {
	var r requests.SetWardrobeRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile wardrobe request: %v", err)
		return
	}

	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId, err := getUserIdFromToken(token)
	if err != nil {
		c.AbortWithStatus(500)
	}
	err = d.db.SetUserWardrobe(userId, r.Wardrobe)
	if err != nil {
		c.AbortWithStatus(500)
		logrus.Errorf("error setting profile mood: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleProfileSetMood(c *gin.Context) {
	var r requests.SetMoodRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile mood request: %v", err)
		return
	}

	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId, err := getUserIdFromToken(token)
	if err != nil {
		c.AbortWithStatus(500)
	}
	err = d.db.SetUserMood(userId, r.Mood)
	if err != nil {
		c.AbortWithStatus(500)
		logrus.Errorf("error setting profile mood: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleProfileUpdateSettings(c *gin.Context) {
	var r requests.UpdateSettingsRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding update profile settings request: %v", err)
		return
	}

	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId, err := getUserIdFromToken(token)
	if err != nil {
		c.AbortWithStatus(500)
	}

	err = d.db.UpdateUserSettings(userId, &models.UserSettings{
		ReceivePushNotifications:  r.ReceivePushNotifications,
		ReceiveEmailNotifications: r.ReceiveEmailNotifications,
	})
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
