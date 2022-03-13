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
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

var (
	FeedPagination = 10
)

func (d *Dutchman) HandleFeed(c *gin.Context) {
	var looks []models.Look

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil {
		logrus.Errorf("error converting page to int: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	offset := page * FeedPagination
	err = d.db.DB().Order("created_at desc").Offset(offset).Limit(FeedPagination).Find(&looks).Error
	if err != nil {
		logrus.Errorf("error getting looks: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"page":  page,
		"looks": looks,
	})
}

func (d *Dutchman) HandleGetLook(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	d.db.DB().First(&look, "slug = ?", slug)

	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(200, look)
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

	d.db.DB().Model(user).Association("DislikedLooks").Delete(&look)

	err = d.db.DB().Model(user).Association("LikedLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

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

	d.db.DB().Model(user).Association("LikedLooks").Delete(&look)

	err = d.db.DB().Model(user).Association("DislikedLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleAddLookToFavorites(c *gin.Context) {
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

	err = d.db.DB().Model(user).Association("FavoriteLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleRemoveLookFromFavorites(c *gin.Context) {
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

	err = d.db.DB().Model(user).Association("FavoriteLooks").Delete(&look)
	if err != nil {
		logrus.Errorf("error adding look to favorites: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandlePinSelection(c *gin.Context) {
	slug := c.Param("selection")

	var selection models.Selection
	d.db.DB().First(&selection, "slug = ?", slug)

	if selection.ID == 0 {
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

	err = d.db.DB().Model(user).Association("PinnedSelections").Append(&selection)
	if err != nil {
		logrus.Errorf("error pinning selection: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleUnpinSelection(c *gin.Context) {
	slug := c.Param("selection")

	var selection models.Selection
	d.db.DB().First(&selection, "slug = ?", slug)

	if selection.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = d.db.DB().Model(user).Association("PinnedSelections").Delete(&selection)
	if err != nil {
		logrus.Errorf("error unpinning selection: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
