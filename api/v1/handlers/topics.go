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

package handlers

import (
	"github.com/gin-gonic/gin"
	database "github.com/lightswitch/papaya-api/pkg/database"
	"github.com/lightswitch/papaya-api/pkg/database/models"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func HandleGetRecommendedTopics(c *gin.Context) {
	var result []models.Topic

	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}

	offset := int(page * 20)
	err := database.DB().Order("created_at desc").Limit(20).Offset(offset).Find(&result).Error
	if err != nil {
		logrus.Errorf("error getting all selections")
		c.AbortWithStatus(500)
	}

	c.JSON(200, result)
}

func HandleGetPopularTopics(c *gin.Context) {
	var result []models.Topic

	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}

	offset := int(page * 20)
	err := database.DB().Order("created_at desc").Limit(20).Offset(offset).Find(&result).Error
	if err != nil {
		logrus.Errorf("error getting all selections")
		c.AbortWithStatus(500)
	}

	c.JSON(200, result)
}

func HandleGetSavedTopics(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var topics []*models.Topic
	err = database.DB().Model(&user).Association("SavedTopics").Find(&topics)
	if err != nil {
		logrus.Errorf("error getting saved topics: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, topics)
}

func HandleGetTopic(c *gin.Context) {
	slug := c.Param("topic")
	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}

	var topic models.Topic
	database.DB().First(&topic, "slug = ?", slug)

	if topic.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	offset := int(page * 20)
	var looks []models.Look

	err := database.DB().Model(topic).Order("created_at desc").Limit(20).Offset(offset).Association("Looks").Find(&looks)
	if err != nil {
		logrus.Errorf("error getting topic looks: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var isSaved bool
	database.DB().Raw("SELECT COUNT(1) FROM saved_topics WHERE user_id = ? AND topic_id = ?", user.ID, topic.ID).Scan(&isSaved)

	c.JSON(200, gin.H{
		"topic":   topic,
		"looks":   looks,
		"isSaved": isSaved,
	})
}

func HandleSaveTopic(c *gin.Context) {
	slug := c.Param("topic")

	var topic models.Topic
	database.DB().First(&topic, "slug = ?", slug)

	if topic.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// user
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = database.DB().Model(user).Association("SavedTopics").Append(&topic)
	if err != nil {
		logrus.Errorf("error watching topic: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleUnsaveTopic(c *gin.Context) {
	slug := c.Param("topic")

	var topic models.Topic
	database.DB().First(&topic, "slug = ?", slug)

	if topic.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// user
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = database.DB().Model(user).Association("SavedTopics").Delete(&topic)
	if err != nil {
		logrus.Errorf("error watching topic: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
