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

package papaya

import (
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/papaya/models"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (d *Dutchman) HandleGetTopics(c *gin.Context) {
	var result []models.Topic

	err := d.db.DB().Find(&result).Error
	if err != nil {
		logrus.Errorf("error getting all selections")
		c.AbortWithStatus(500)
	}

	c.JSON(200, result)
}

func (d *Dutchman) HandleGetTopic(c *gin.Context) {
	slug := c.Param("topic")

	var topic models.Topic
	d.db.DB().First(&topic, "slug = ?", slug)

	if topic.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(200, topic)
}

func (d *Dutchman) HandleWatchTopic(c *gin.Context) {
	slug := c.Param("topic")

	var topic models.Topic
	d.db.DB().First(&topic, "slug = ?", slug)

	if topic.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// user

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = d.db.DB().Model(user).Association("Topics").Append(&topic)
	if err != nil {
		logrus.Errorf("error watching topic: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleUnwatchTopic(c *gin.Context) {
	slug := c.Param("topic")

	var topic models.Topic
	d.db.DB().First(&topic, "slug = ?", slug)

	if topic.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = d.db.DB().Model(user).Association("Topics").Delete(&topic)
	if err != nil {
		logrus.Errorf("error unwatching topic: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
