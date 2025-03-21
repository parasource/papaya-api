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
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/parasource/papaya-api/pkg/gorse"
	"github.com/rs/zerolog/log"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

const (
	searchSql = `SELECT searches_%[1]v.*, 
		ts_rank(searches_%[1]v.tsv, plainto_tsquery('russian', ?)) as rank 
		FROM searches_%[1]v
		LEFT JOIN look_items li on searches_%[1]v.id = li.look_id JOIN wardrobe_items wi on wi.id = li.wardrobe_item_id
		JOIN (
            VALUES %[2]v
        ) AS x (id, ordering) ON wi.id = x.id
		WHERE searches_%[1]v.tsv @@ plainto_tsquery('russian', ?)
		OR wi.id IN (?)
		ORDER BY x.ordering DESC, rank desc
		OFFSET ? LIMIT ?`

	searchSqlNoWardrobeFound = `SELECT searches_%[1]v.*, 
		ts_rank(searches_%[1]v.tsv, plainto_tsquery('russian', ?)) as rank 
		FROM searches_%[1]v
		LEFT JOIN look_items li on searches_%[1]v.id = li.look_id JOIN wardrobe_items wi on wi.id = li.wardrobe_item_id
		WHERE searches_%[1]v.tsv @@ plainto_tsquery('russian', ?)
		ORDER BY rank desc
		OFFSET ? LIMIT ?`
)

type SearchDBWardrobe struct {
	ID   int     `json:"id"`
	Rank float32 `json:"rank"`
}

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

	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	if len(params["q"]) != 1 {
		c.AbortWithStatus(400)
		return
	}
	searchQuery := params["q"][0]
	if searchQuery == "" {
		c.JSON(http.StatusNoContent, []int{})
		return
	}

	sr := models.SearchRecord{
		Query:   searchQuery,
		UserID:  user.ID,
		Visible: true,
	}
	err = database.DB().Create(&sr).Error
	if err != nil {
		logrus.Errorf("error recording user search: %v", err)
	}

	var page int64
	if _, ok := params["page"]; !ok {
		page = 0
	} else {
		page, _ = strconv.ParseInt(params["page"][0], 10, 64)
	}
	offset := int(page * 20)

	var res []*SearchDBResult

	// First we need to query wardrobe matches,
	// as it is our main goal
	sqlQueryWardrobe := `select wardrobe_items.id, ts_rank(tsv, plainto_tsquery('pg_catalog.russian', ?)) as rank
from wardrobe_items
where tsv @@ plainto_tsquery('pg_catalog.russian', ?)
order by rank desc limit 5;`
	var wardrobeSearchResult []SearchDBWardrobe
	err = database.DB().Debug().Raw(sqlQueryWardrobe, searchQuery, searchQuery).Scan(&wardrobeSearchResult).Error
	if err != nil {
		log.Error().Err(err).Msg("error querying wardrobe")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	var wardrobeIds []int
	for _, item := range wardrobeSearchResult {
		wardrobeIds = append(wardrobeIds, item.ID)
	}

	dbQuery := fmt.Sprintf(searchSql, user.Sex, idsToInClauseWithOrdering(wardrobeIds))
	if len(wardrobeIds) == 0 {
		dbQuery = fmt.Sprintf(searchSqlNoWardrobeFound, user.Sex)
	}

	if len(wardrobeIds) > 0 {
		err = database.DB().Debug().Raw(dbQuery, searchQuery, searchQuery, wardrobeIds, offset, 20).Find(&res).Error
	} else {
		err = database.DB().Debug().Raw(dbQuery, searchQuery, searchQuery, offset, 20).Find(&res).Error
	}
	if err != nil {
		logrus.Errorf("error searching: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
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
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	err = database.DB().Find(&topics, topicIDs).Error
	if err != nil {
		logrus.Errorf("error finding topics by ids: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Filling out ranks for models
	for _, look := range looks {
		look.Rank = ranks[look.ID]
	}
	for _, topic := range topics {
		topic.Rank = ranks[topic.ID]
	}

	c.JSON(200, gin.H{
		"looks":  looks,
		"topics": topics,
	})
}

func idsToInClauseWithOrdering(ids []int) string {
	clause := ""
	for i := 0; i < len(ids); i++ {
		if i == len(ids)-1 {
			clause += fmt.Sprintf("(%v, %v)", ids[i], i+1)
			break
		}
		clause += fmt.Sprintf("(%v, %v), ", ids[i], i+1)
	}
	return clause
}

func HandleSearchSuggestions(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var sr []*models.SearchRecord
	err = database.DB().Where("user_id = ?", user.ID).Where("visible = ?", true).Order("id desc").Limit(10).Find(&sr).Error
	if err != nil {
		logrus.Errorf("error getting search records: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
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

	lookSlugs, err := gorse.RecommendPopular(user.Sex, 10)
	if err != nil {
		logrus.Errorf("error getting popular looks: %v", err)
		c.AbortWithStatus(500)
		return
	}

	if len(lookSlugs) == 0 {
		err = database.DB().Order("RANDOM()").Where("sex = ?", user.Sex).Limit(10).Find(&looks).Error
	} else {
		err = database.DB().Debug().Where("slug in ?", lookSlugs).Find(&looks).Error
		if err != nil {
			logrus.Errorf("error getting popular looks from db: %v", err)
			c.AbortWithStatus(500)
			return
		}
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

func HandleSearchClearHistory(c *gin.Context) {
	user, err := GetUser(c)
	if err != nil {
		logrus.Errorf("error getting user: %v", err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = database.DB().Model(&models.SearchRecord{}).Where("user_id = ?", user.ID).Update("visible", false).Error
	if err != nil {
		logrus.Errorf("error clearing user search history: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(200, []struct{}{})
}

func HandleSearchAutofill(c *gin.Context) {
	params := c.Request.URL.Query()
	q := params["q"]
	if q[0] == "" {
		c.JSON(http.StatusNoContent, []int{})
		return
	}

	var sr []*models.SearchRecord
	err := database.DB().Raw("select query, count(id) as freq from search_records where query like ? group by query order by freq desc limit ?", q[0]+"%", 10).Find(&sr).Error
	if err != nil {
		logrus.Errorf("erorr searching: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, sr)
}
