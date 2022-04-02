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
	"strconv"
)

func (d *Dutchman) HandleGetTopics(c *gin.Context) {
	var result []models.Topic

	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}

	offset := int(page * 20)
	err := d.db.DB().Order("created_at desc").Limit(20).Offset(offset).Find(&result).Error
	if err != nil {
		logrus.Errorf("error getting all selections")
		c.AbortWithStatus(500)
	}

	c.JSON(200, result)
}

func (d *Dutchman) HandleGetTopic(c *gin.Context) {
	slug := c.Param("topic")
	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}

	var topic models.Topic
	d.db.DB().First(&topic, "slug = ?", slug)

	if topic.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	offset := int(page * 20)
	var looks []models.Look

	err := d.db.DB().Model(topic).Order("created_at desc").Limit(20).Offset(offset).Association("Looks").Find(&looks)
	if err != nil {
		logrus.Errorf("error getting topic looks: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var isWatched bool
	d.db.DB().Raw("SELECT COUNT(1) FROM watched_topics WHERE user_id = ? AND topic_id = ?", user.ID, topic.ID).Scan(&isWatched)

	c.JSON(200, gin.H{
		"topic":     topic,
		"looks":     looks,
		"isWatched": isWatched,
	})
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
