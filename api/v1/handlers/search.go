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
	"github.com/sirupsen/logrus"
	"strconv"
)

type SearchDBResult struct {
	OriginTable string  `json:"origin_table"`
	ID          uint    `json:"id"`
	Tsv         string  `json:"tsv" gorm:"type:tsvector"`
	Rank        float32 `json:"rank"`
}

type SearchSuggestion struct {
	Query string `json:"query"`
}

func HandleSearch(c *gin.Context) {

	params := c.Request.URL.Query()

	q := params["q"]
	if q[0] == "" {
		c.JSON(204, []int{})
		return
	}

	//user, err := GetUser(c)
	//if err != nil {
	//	logrus.Errorf("error getting user: %v", err)
	//	c.AbortWithStatus(403)
	//	return
	//}

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}
	offset := int(page * 20)

	var res []*SearchDBResult

	err := database.DB().Raw("SELECT searches.*, ts_rank(searches.tsv, plainto_tsquery('russian', ?)) as rank FROM searches WHERE searches.tsv @@ plainto_tsquery('russian', ?) OFFSET ? LIMIT ?", q[0], q[0], offset, 20).Find(&res).Error
	if err != nil {
		logrus.Errorf("erorr searching: %v", err)
		c.AbortWithStatus(500)
		return
	}

	if len(res) == 0 {
		c.JSON(200, gin.H{
			"looks":  []int{},
			"topics": []int{},
		})
		return
	}

	lookIDs := make([]uint, len(res))
	topicIDs := make([]uint, len(res))

	ranks := make(map[uint]float32, len(res))
	var l, t = 0, 0
	for _, result := range res {
		if result.OriginTable == "looks" {
			lookIDs[l] = result.ID
			l++
		} else if result.OriginTable == "topics" {
			topicIDs[t] = result.ID
			t++
		}
		// Recording rank score
		ranks[result.ID] = result.Rank
	}

	var looks []*models.Look
	var topics []*models.Topic

	err = database.DB().Find(&looks, lookIDs).Error
	if err != nil {
		logrus.Errorf("error finding looks by ids: %v", err)
		c.AbortWithStatus(500)
		return
	}
	err = database.DB().Find(&topics, topicIDs).Error
	if err != nil {
		logrus.Errorf("error finding topics by ids: %v", err)
		c.AbortWithStatus(500)
		return
	}

	// Filling out ranks for models
	for _, look := range looks {
		look.Rank = ranks[look.ID]
	}
	for _, topic := range topics {
		topic.Rank = ranks[topic.ID]
	}

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}
	sr := models.SearchRecord{
		Query:  q[0],
		UserID: user.ID,
	}
	err = database.DB().Create(&sr).Error
	if err != nil {
		logrus.Errorf("error recording user search: %v", err)
	}

	c.JSON(200, gin.H{
		"looks":  looks,
		"topics": topics,
	})
}

func HandleSearchSuggestions(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(403)
		return
	}

	var sr []*models.SearchRecord
	err = database.DB().Where("user_id = ?", user.ID).Order("id desc").Limit(10).Find(&sr).Error
	if err != nil {
		logrus.Errorf("error getting search records: %v", err)
		c.AbortWithStatus(500)
		return
	}

	var suggestions []*SearchSuggestion

	err = database.DB().Raw(`SELECT query, count(id) AS c FROM search_records
                             WHERE created_at >= NOW() - interval '7 day'
                             GROUP BY search_records.query ORDER BY c DESC;`).Find(&suggestions).Error
	if err != nil {
		logrus.Errorf("error getting search suggestions: %v", err)
	}

	var looks []*models.Look
	var topics []*models.Topic

	err = database.DB().Order("RANDOM()").Limit(10).Find(&looks).Error
	if err != nil {
		logrus.Errorf("error getting popular looks: %v", err)
		c.AbortWithStatus(500)
		return
	}

	err = database.DB().Order("RANDOM()").Limit(10).Find(&topics).Error
	if err != nil {
		logrus.Errorf("error getting popular topics: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"search": gin.H{
			"history":     sr,
			"suggestions": suggestions,
		},
		"looks":  looks,
		"topics": topics,
	})
}

func HandleSearchAutofill(c *gin.Context) {
	params := c.Request.URL.Query()
	q := params["q"]
	if q[0] == "" {
		c.JSON(204, []int{})
		return
	}

	var sr []*models.SearchRecord
	err := database.DB().Raw("SELECT search_records.*, ts_rank(search_records.tsv, plainto_tsquery('russian', ?)) as rank FROM search_records WHERE search_records.tsv @@ plainto_tsquery('russian', ?) LIMIT ?", q[0], q[0], 10).Find(&sr).Error
	if err != nil {
		logrus.Errorf("erorr searching: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, sr)
}
