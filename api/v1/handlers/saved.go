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
	"github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/parasource/papaya-api/pkg/gorse"
	"github.com/sirupsen/logrus"
	"strconv"
)

func HandleSaved(c *gin.Context) {
	var result []models.Look

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}
	database.DB().Model(&user).Association("SavedLooks").Find(&result)

	c.JSON(200, result)
}

func HandleSavedAdd(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	err := database.DB().Where("slug = ?", slug).Find(&look).Error
	if err != nil {
		c.AbortWithStatus(404)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = database.DB().Model(&user).Association("SavedLooks").Append(&look)
	if err != nil {
		logrus.Errorf("error adding look to saved: %v", err)
	}

	err = gorse.Star(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("gorse error starring look: %v", err)
	}

	c.Status(200)
}

func HandleSavedRemove(c *gin.Context) {
	slug := c.Param("look")

	var look models.Look
	err := database.DB().Where("slug = ?", slug).Find(&look).Error
	if err != nil {
		c.AbortWithStatus(404)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	err = database.DB().Model(&user).Association("SavedLooks").Delete(&look)
	if err != nil {
		logrus.Errorf("error removing look from saved: %v", err)
	}

	err = gorse.Unstar(strconv.Itoa(int(user.ID)), strconv.Itoa(int(look.ID)))
	if err != nil {
		logrus.Errorf("gorse error starring look: %v", err)
	}

	c.Status(200)
}
