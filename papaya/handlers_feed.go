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
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

var (
	FeedPagination = 10
)

func (d *Papaya) HandleFeed(c *gin.Context) {
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

	var styles []models.Style
	err = d.db.DB().Preload("Looks.Items").Find(&styles).Error
	if err != nil {
		logrus.Errorf("error getting styles: %v", err)
		c.AbortWithStatus(500)
		return
	}

	var todayLookId int
	d.db.DB().Raw("SELECT look_id FROM today_looks WHERE user_id = ? LIMIT 1", user.ID).Scan(&todayLookId)

	var todayLook models.Look
	d.db.DB().Preload("Items").First(&todayLook, "id = ?", todayLookId)

	err = d.db.DB().Model(&user).Association("TodayLook").Find(&todayLook)
	if err != nil {
		logrus.Errorf("error getting today's look: %v", err)
		c.AbortWithStatus(500)
		return
	}

	result := gin.H{
		"todayLook": todayLook,
		"page":      page,
		"looks":     looks,
		"styles":    styles,
	}
	c.JSON(200, result)
}

func (d *Papaya) HandleFeedByStyle(c *gin.Context) {
	slug := c.Param("style")

	params := c.Request.URL.Query()

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}

	offset := int(page * 20)

	var style models.Style
	d.db.DB().Preload("Looks.Items").First(&style, "slug = ?", slug)

	if style.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var looks []*models.Look
	err := d.db.DB().
		Raw("SELECT * FROM looks JOIN look_styles ls on looks.id = ls.look_id WHERE ls.style_id = ?", style.ID).
		Order("created_at desc").
		Offset(offset).Preload("Items.Urls.Brand").
		Limit(FeedPagination).Find(&looks).Error
	if err != nil {
		logrus.Errorf("error getting feed looks by style: %v", err)
	}

	c.JSON(200, gin.H{
		"looks": looks,
		"style": style,
	})
}

func (d *Papaya) HandleGetLook(c *gin.Context) {
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

func (d *Papaya) HandleGetLookItem(c *gin.Context) {
	slugLook := c.Param("look")
	var look models.Look
	d.db.DB().First(&look, "slug = ?", slugLook)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	itemId := c.Param("item")
	var item models.WardrobeItem
	d.db.DB().Preload("Urls.Brand").First(&item, "id = ?", itemId)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	var looks []models.Look
	d.db.DB().Raw("SELECT looks.* FROM looks JOIN look_items li on looks.id = li.look_id WHERE li.look_id = ? AND looks.id != ? LIMIT 10", item.ID, look.ID).Scan(&looks)

	c.JSON(200, gin.H{
		"item":  item,
		"looks": looks,
	})
}

func (d *Papaya) HandleLikeLook(c *gin.Context) {
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

func (d *Papaya) HandleUnlikeLook(c *gin.Context) {
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

func (d *Papaya) HandleDislikeLook(c *gin.Context) {
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

func (d *Papaya) HandleUndislikeLook(c *gin.Context) {
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

func (d *Papaya) GetLikedLooks(c *gin.Context) {
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
