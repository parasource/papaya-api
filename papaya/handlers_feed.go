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

var (
	FeedPagination = 10
)

func (d *Dutchman) HandleFeed(c *gin.Context) {
	var looks []models.Look

	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}

	offset := int(page * 20)
	err := d.db.DB().Order("created_at desc").Offset(offset).Preload("Items.Urls.Brand").Limit(FeedPagination).Find(&looks).Error
	if err != nil {
		logrus.Errorf("error getting looks: %v", err)
		c.AbortWithStatus(500)
		return
	}

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var topics []models.Topic
	err = d.db.DB().Model(&user).Association("Topics").Find(&topics)
	if err != nil {
		logrus.Errorf("error getting watched topics: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"page":   page,
		"looks":  looks,
		"topics": topics,
	})
}

func (d *Dutchman) HandleGetLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	d.db.DB().Preload("Items.Urls.Brand").First(&look, "slug = ?", slug)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = d.adviser.Read(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("error submitting 'read' feedback to adviser: %v", err)
	}

	var isLiked bool
	d.db.DB().Raw("SELECT COUNT(1) FROM liked_looks WHERE user_id = ? AND look_id = ?", user.ID, look.ID).Scan(&isLiked)

	var isDisliked bool
	d.db.DB().Raw("SELECT COUNT(1) FROM disliked_looks WHERE user_id = ? AND look_id = ?", user.ID, look.ID).Scan(&isDisliked)

	c.JSON(200, gin.H{
		"look":       look,
		"isLiked":    isLiked,
		"isDisliked": isDisliked,
	})
}

func (d *Dutchman) HandleLikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	d.db.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
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

	err = d.db.DB().Model(user).Association("LikedLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	err = d.adviser.Like(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("error submitting 'like' feedback to adviser: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleUnlikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	d.db.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
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

	d.db.DB().Model(user).Association("LikedLooks").Delete(&look)

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleDislikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	d.db.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
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

	err = d.db.DB().Model(user).Association("DislikedLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleUndislikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	d.db.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
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

	err = d.db.DB().Model(user).Association("DislikedLooks").Delete(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) GetLikedLooks(c *gin.Context) {
	// user

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var looks []models.Look
	err = d.db.DB().Model(user).Association("LikedLooks").Find(&looks)
	if err != nil {
		logrus.Errorf("error getting liked looks: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, looks)
}
