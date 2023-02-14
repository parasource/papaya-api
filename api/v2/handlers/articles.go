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

func HandleGetArticle(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.AbortWithStatus(404)
		return
	}

	var article models.Article
	err := database.DB().Where("slug = ?", slug).Find(&article).Error
	if err != nil {
		logrus.Errorf("error getting articles: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, article)
}
