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
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"math"
	"net/http"
	"strconv"
)

func HandleGetArticles(c *gin.Context) {
	var err error
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
	offset := 4 + int(page*8)

	var pinned, articles []models.Article
	err = database.DB().Raw(`select * from articles
         where deleted_at is null
         order by id desc limit 4 offset 0`).Scan(&pinned).Error
	if err != nil {
		log.Error().Err(err).Msg("error getting pinned articles")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = database.DB().Raw(`select * from articles 
         where deleted_at is null 
         order by id desc
         limit 8 offset ?`, offset).Scan(&articles).Error
	if err != nil {
		log.Error().Err(err).Msg("error getting articles")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var articlesCount int
	err = database.DB().
		Raw("select count(id) from articles where deleted_at is null").Scan(&articlesCount).Error
	if err != nil {
		log.Error().Err(err).Msg("error getting articles count")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, gin.H{
		"pinned":     pinned,
		"articles":   articles,
		"pagesCount": math.Ceil(float64(articlesCount-4) / 8),
	})
}

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
