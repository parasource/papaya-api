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
	"github.com/parasource/papaya-api/pkg/adviser"
	database "github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/parasource/papaya-api/pkg/gorse"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

var (
	FeedPagination = 20
)

func HandleFeed(c *gin.Context) {

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, err = strconv.ParseInt(params["page"][0], 10, 64)
		if err != nil {
			c.AbortWithStatus(400)
			return
		}
	}

	// Feed looks
	looks, err := adviser.Get().Feed(user, int(page))
	if err != nil {
		logrus.Errorf("error getting feed: %v", err)
		c.AbortWithStatus(500)
		return
	}

	// Categories
	var categories []models.Category
	err = database.DB().Find(&categories).Error
	if err != nil {
		logrus.Errorf("error getting categories: %v", err)
		c.AbortWithStatus(500)
		return
	}

	// Topics
	var topics []models.Topic
	err = database.DB().Order("RANDOM()").Limit(10).Find(&topics).Error
	if err != nil {
		logrus.Errorf("error getting popular topics: %v", err)
		c.AbortWithStatus(500)
		return
	}

	// Articles
	var articles []models.Article
	err = database.DB().Where("sex = ?", user.Sex).Order("RANDOM()").Limit(3).Find(&articles).Error
	if err != nil {
		logrus.Errorf("error getting articles: %v", err)
		c.AbortWithStatus(500)
		return
	}

	// Alerts
	var alerts []models.Alert
	err = database.DB().Find(&alerts).Error
	if err != nil {
		log.Error().Err(err).Msg("error getting alerts")
	}

	result := gin.H{
		"page":               page,
		"topics":             topics,
		"looks":              looks,
		"categories":         categories,
		"articles":           articles,
		"alerts":             alerts,
		"latest_app_version": "1.0.3",
	}
	c.JSON(200, result)
}

func HandleFeedByCategory(c *gin.Context) {
	var err error
	slug := c.Param("category")

	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, err = strconv.ParseInt(params["page"][0], 10, 64)
		if err != nil {
			c.AbortWithStatus(400)
			return
		}
	}

	offset := int(page * 20)

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var category models.Category
	database.DB().First(&category, "slug = ?", slug)

	if category.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var looks []*models.Look
	err = database.DB().Debug().
		Raw("SELECT * FROM looks JOIN look_categories lc on looks.id = lc.look_id WHERE looks.deleted_at IS NULL AND lc.category_id = ? AND looks.sex = ?", category.ID, user.Sex).
		Order("id DESC").
		Offset(offset).Preload("Items.Urls.Brand").
		Limit(FeedPagination).Find(&looks).Error
	if err != nil {
		logrus.Errorf("error getting feed looks by style: %v", err)
	}

	c.JSON(200, gin.H{
		"looks":    looks,
		"category": category,
		"page":     page,
	})
}

func HandleGetLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	database.DB().Preload("Items.Urls.Brand").Preload("Categories").First(&look, "slug = ?", slug)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = gorse.Read(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("error submitting 'read' feedback to adviser: %v", err)
	}

	var isLiked bool
	database.DB().Raw("SELECT COUNT(1) FROM liked_looks WHERE user_id = ? AND look_id = ?", user.ID, look.ID).Scan(&isLiked)

	var isDisliked bool
	database.DB().Raw("SELECT COUNT(1) FROM disliked_looks WHERE user_id = ? AND look_id = ?", user.ID, look.ID).Scan(&isDisliked)

	var isSaved bool
	database.DB().Raw("SELECT COUNT(1) FROM saved_looks WHERE user_id = ? AND look_id = ?", user.ID, look.ID).Scan(&isSaved)

	var similar []models.Look
	err = database.DB().Where("sex = ?", user.Sex).Limit(8).Order("random()").Find(&similar).Error
	if err != nil {
		log.Error().Err(err).Msg("error finding similar looks")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"look":       look,
		"isLiked":    isLiked,
		"isDisliked": isDisliked,
		"isSaved":    isSaved,
		"similar":    similar,
	})
}

func HandleGetLookItem(c *gin.Context) {
	slugLook := c.Param("look")
	var look models.Look
	database.DB().First(&look, "slug = ?", slugLook)

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	itemId := c.Param("item")
	var item models.WardrobeItem
	database.DB().Preload("Urls.Brand").First(&item, "id = ?", itemId)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var looks []models.Look
	database.DB().Raw("SELECT looks.* FROM looks JOIN look_items li on looks.id = li.look_id WHERE looks.sex = ? AND li.wardrobe_item_id = ? AND looks.id != ? LIMIT 20", user.Sex, item.ID, look.ID).Scan(&looks)

	c.JSON(200, gin.H{
		"item":  item,
		"looks": looks,
	})
}

func HandleLikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	database.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = database.DB().Model(user).Association("LikedLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	err = gorse.Like(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("error submitting 'like' feedback to adviser: %v", err)
	}
	err = gorse.Undislike(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("gorse error disliking look: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleUnlikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	database.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	database.DB().Model(user).Association("LikedLooks").Delete(&look)

	err = gorse.Unlike(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("gorse error unliking look: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleDislikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	database.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
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

	err = database.DB().Model(user).Association("DislikedLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	err = gorse.Unlike(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("gorse error unliking look: %v", err)
	}
	err = gorse.Dislike(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("gorse error disliking look: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleUndislikeLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	database.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = database.DB().Model(user).Association("DislikedLooks").Delete(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	err = gorse.Undislike(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("gorse error undisliking look: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func GetLikedLooks(c *gin.Context) {

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var looks []models.Look
	err = database.DB().Model(user).Association("LikedLooks").Find(&looks)
	if err != nil {
		logrus.Errorf("error getting liked looks: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, looks)
}
