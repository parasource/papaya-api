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
	"github.com/lightswitch/papaya-api/pkg/db"
	models2 "github.com/lightswitch/papaya-api/pkg/db/models"
	"github.com/sirupsen/logrus"
	"strconv"
)

type SearchDBResult struct {
	OriginTable string  `json:"origin_table"`
	ID          uint    `json:"id"`
	Tsv         string  `json:"tsv" gorm:"type:tsvector"`
	Rank        float32 `json:"rank"`
}

func HandleSearch(c *gin.Context) {

	params := c.Request.URL.Query()

	q := params["q"]
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

	var looks []*models2.Look
	var topics []*models2.Topic

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
	sr := models2.SearchRecord{
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
