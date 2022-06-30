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
	"github.com/lightswitch/papaya-api/api/v1/requests"
	database "github.com/lightswitch/papaya-api/pkg/database"
	"github.com/lightswitch/papaya-api/pkg/database/models"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func HandleGetCollections(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var collections []*models.Collection
	err = database.DB().Model(&user).Preload("Looks").Association("Collections").Find(&collections)
	if err != nil {
		logrus.Errorf("error getting user's collections: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, collections)
}

func HandleCreateCollection(c *gin.Context) {
	var r requests.CreateCollectionRequest
	err := c.BindJSON(&r)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	coll := &models.Collection{
		Name:   r.Name,
		UserID: user.ID,
	}
	database.DB().Create(coll)
	if err = database.DB().Model(user).Association("Collections").Append(coll); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"coll":    coll,
	})
}

func HandleGetCollection(c *gin.Context) {
	collID, _ := strconv.Atoi(c.Param("collection"))

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var coll models.Collection
	database.DB().Preload("Looks").First(&coll, "id = ?", collID)
	if coll.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if coll.UserID != user.ID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.JSON(200, coll)
}

func HandleDeleteCollection(c *gin.Context) {
	collID, _ := strconv.Atoi(c.Param("collection"))

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var coll models.Collection
	database.DB().First(&coll, "id = ?", collID)
	if coll.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if coll.UserID != user.ID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = database.DB().Model(&coll).Association("Looks").Delete(coll.Looks)
	if err != nil {
		logrus.Errorf("error removing looks from collection before deleting: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	database.DB().Delete(&coll)

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleCollectionAddLook(c *gin.Context) {
	collID, _ := strconv.Atoi(c.Param("collection"))
	lookSlug := c.Param("look")

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var coll models.Collection
	database.DB().First(&coll, "id = ?", collID)
	if coll.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if coll.UserID != user.ID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var look models.Look
	database.DB().First(&look, "slug = ?", lookSlug)
	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = database.DB().Model(&coll).Association("Looks").Append(&look)
	if err != nil {
		logrus.Errorf("error associating look with collection: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	//p.database.DB().Save()

	c.JSON(200, gin.H{
		"success": true,
	})
}

func HandleCollectionRemoveLook(c *gin.Context) {
	collID, _ := strconv.Atoi(c.Param("collection"))
	lookSlug := c.Param("look")

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var coll models.Collection
	database.DB().First(&coll, "id = ?", collID)
	if coll.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	if coll.UserID != user.ID {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var look models.Look
	database.DB().First(&look, "slug = ?", lookSlug)
	if look.ID == 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	err = database.DB().Model(&coll).Association("Looks").Delete(&look)
	if err != nil {
		logrus.Errorf("error associating look with collection: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
