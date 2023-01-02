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
	"github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/sirupsen/logrus"
)

func HandleGetWardrobeCategories(c *gin.Context) {
	var categories []*models.WardrobeCategory

	err := database.DB().Find(&categories).Error
	if err != nil {
		logrus.Errorf("error getting all wardrobe categories: %v", err)
		c.AbortWithStatus(500)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var preview string
	var count int

	for _, category := range categories {
		err = database.DB().Raw("SELECT image FROM wardrobe_items WHERE wardrobe_category_id = ? AND (sex = ? OR sex = 'unisex') LIMIT 1", category.ID, user.Sex).Scan(&preview).Error
		if err != nil {
			logrus.Errorf("error getting preview for wardrobe category: %v", err)
		}
		category.Preview = preview

		err = database.DB().Raw("SELECT COUNT(*) AS count FROM wardrobe_items WHERE wardrobe_category_id = ? AND (sex = ? OR sex = 'unisex')", category.ID, user.Sex).Scan(&count).Error
		if err != nil {
			logrus.Errorf("error getting items count for wardrobe category: %v", err)
		}
		category.ItemsCount = count
	}

	c.JSON(200, categories)
}

func HandleGetWardrobeItems(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.AbortWithStatus(400)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var items []*models.WardrobeItem
	err = database.DB().Raw("select * from wardrobe_items where wardrobe_category_id = ? AND (sex = ? OR sex = 'unisex')", category, user.Sex).Find(&items).Error
	//err = database.DB().Where("wardrobe_category_id = ?", category).Where("sex", user.Sex).Find(&items).Error
	if err != nil {
		logrus.Errorf("error getting wardrobe items: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, items)
}
