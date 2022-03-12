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
	"github.com/lightswitch/dutchman-backend/dutchman/requests"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

func (d *Dutchman) HandleCreateCollection(c *gin.Context) {
	var r requests.CreateCollectionRequest
	err := c.BindJSON(&r)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	coll := &models.Collection{
		Name:   r.Name,
		UserID: user.ID,
	}
	d.db.DB().Create(coll)
	if err = d.db.DB().Model(user).Association("Collections").Append(coll); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"coll":    coll,
	})
}

func (d *Dutchman) HandleGetCollection(c *gin.Context) {
	collID, _ := strconv.Atoi(c.Param("collection"))

	user, err := d.GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var coll models.Collection
	d.db.DB().First(&coll, "id = ?", collID)
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
