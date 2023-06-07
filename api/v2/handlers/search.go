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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"net/http"
	"strconv"
	"strings"
)

const (
	searchSql = `SELECT looks.*,
        ts_rank(looks.tsv, plainto_tsquery('russian', ?)) as rank
FROM looks
		LEFT JOIN look_items li on looks.id = li.look_id JOIN wardrobe_items wi on wi.id = li.wardrobe_item_id
    JOIN (
        VALUES %[1]v
    ) AS x (id, ordering) ON wi.id = x.id
	WHERE (looks.tsv @@ plainto_tsquery('russian', ?)
   	OR wi.id IN (?))
	AND looks.sex = ?
	ORDER BY x.ordering DESC, rank desc
	OFFSET 0 LIMIT 10
`

	searchSqlNoWardrobeFound = `SELECT looks.*,
        ts_rank(looks.tsv, plainto_tsquery('russian', ?)) as rank
FROM looks
		LEFT JOIN look_items li on looks.id = li.look_id JOIN wardrobe_items wi on wi.id = li.wardrobe_item_id
		WHERE looks.tsv @@ plainto_tsquery('russian', ?)
		AND looks.sex = ?
		GROUP BY looks.id
		ORDER BY rank desc
		OFFSET ? LIMIT ?
`

	searchSqlWardrobe = `SELECT wardrobe_items.id, ts_rank(wardrobe_items.tsv, plainto_tsquery('pg_catalog.russian', ?)) AS rank,
       count(li.id) AS items_count
    FROM wardrobe_items JOIN look_items li on wardrobe_items.id = li.wardrobe_item_id
    WHERE wardrobe_items.tsv @@ plainto_tsquery('pg_catalog.russian', ?) and sex = ?
    GROUP BY wardrobe_items.id, rank
	ORDER BY rank, items_count DESC LIMIT 5;`
)

type SearchSuggestion struct {
	Query string `json:"query"`
}

type SearchDBWardrobe struct {
	ID   int     `json:"id"`
	Rank float32 `json:"rank"`
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
		Query:   strings.ToLower(searchQuery),
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

	// First we need to query wardrobe matches,
	// as it is our main goal
	var wardrobeSearchResult []SearchDBWardrobe
	err = database.DB().Debug().Raw(searchSqlWardrobe, searchQuery, searchQuery, user.Sex).Scan(&wardrobeSearchResult).Error
	if err != nil {
		log.Error().Err(err).Msg("error querying wardrobe")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	var wardrobeIds []int
	for _, item := range wardrobeSearchResult {
		wardrobeIds = append(wardrobeIds, item.ID)
	}

	var wardrobeItems []models.WardrobeItem
	if len(wardrobeIds) > 0 {
		err = database.DB().Where("id", wardrobeIds).Preload("WardrobeCategory").Preload("Urls.Brand").Find(&wardrobeItems).Error
		if err != nil {
			log.Error().Err(err).Msg("error searching wardrobe items")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		log.Warn().Str("query", searchQuery).Msg("did not found any wardrobe items")
	}

	dbQuery := fmt.Sprintf(searchSql, idsToInClauseWithOrdering(wardrobeIds))
	if len(wardrobeIds) == 0 {
		dbQuery = searchSqlNoWardrobeFound
	}

	var looks []models.Look
	if len(wardrobeIds) > 0 {
		err = database.DB().Debug().Raw(dbQuery, searchQuery, searchQuery, wardrobeIds, user.Sex, offset, 20).Find(&looks).Error
	} else {
		err = database.DB().Debug().Raw(dbQuery, searchQuery, searchQuery, user.Sex, offset, 20).Find(&looks).Error
	}
	if err != nil {
		logrus.Errorf("error searching: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if len(looks) == 0 {
		c.JSON(200, gin.H{
			"looks":          []interface{}{},
			"wardrobe_items": wardrobeItems,
		})
		return
	}

	c.JSON(200, gin.H{
		"looks":          looks,
		"wardrobe_items": wardrobeItems,
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
	err = database.DB().Where("user_id = ?", user.ID).Where("visible = ?", true).Order("id desc").Limit(5).Find(&sr).Error
	if err != nil {
		logrus.Errorf("error getting search records: %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var suggestions []*SearchSuggestion

	err = database.DB().Raw(`SELECT query, count(id) AS c FROM search_records
                             WHERE created_at >= NOW() - interval '7 day'
                             GROUP BY search_records.query ORDER BY c DESC LIMIT ?;`, 5).Find(&suggestions).Error
	if err != nil {
		logrus.Errorf("error getting search suggestions: %v", err)
	}

	var looks []*models.Look

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

	c.JSON(200, gin.H{
		"search": gin.H{
			"history":     sr,
			"suggestions": suggestions,
		},
		"looks": looks,
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
	query := strings.ToLower(strings.TrimSpace(q[0]))

	// Splitting by tags
	tmpQueryWords := []string{}
	queryWords := strings.Split(query, " ")
	for i := 0; i < len(queryWords); i++ {
		var word string
		if isSeparator(queryWords[i]) {
			word = strings.Join([]string{queryWords[i], queryWords[i+1]}, " ")
			i++
		} else {
			word = queryWords[i]
		}
		tmpQueryWords = append(tmpQueryWords, word)
	}
	queryWordCount := len(tmpQueryWords)

	// Capitalizing first letter
	capitalizedFirstWord := cases.Title(language.Russian).String(tmpQueryWords[0])
	var queryWardrobe string
	if len(tmpQueryWords) > 1 {
		queryWardrobe = capitalizedFirstWord + " " + strings.Join(tmpQueryWords[1:], " ")
		tmpQueryWords[0] = capitalizedFirstWord
	} else {
		queryWardrobe = capitalizedFirstWord
		tmpQueryWords[0] = capitalizedFirstWord
	}

	var wsr []*models.WardrobeItem
	err := database.DB().Raw("select * from wardrobe_items where name like ? limit ?", queryWardrobe+"%", 10).Find(&wsr).Error
	if err != nil {
		logrus.Errorf("error searching: %v", err)
		c.AbortWithStatus(500)
		return
	}

	var sr []*models.SearchRecord
	err = database.DB().Raw("select query, count(id) as freq from search_records where query like ? group by query order by freq desc limit ?", query+"%", 10).Find(&sr).Error
	if err != nil {
		logrus.Errorf("error searching: %v", err)
		c.AbortWithStatus(500)
		return
	}

	tags := []string{}
	counter := 0
	for {
		if counter >= len(wsr) {
			break
		}
		if len(tags) == 3 {
			break
		}

		arr := strings.Split(wsr[counter].Name, " ")
		if len(arr)-queryWordCount < 1 {
			counter++
			continue
		}

		tmpQueryTags := []string{}
		for i := 0; i < len(arr); i++ {
			var word string
			if isSeparator(arr[i]) {
				word = strings.Join([]string{arr[i], arr[i+1]}, " ")
				i++
			} else {
				word = arr[i]
			}
			tmpQueryTags = append(tmpQueryTags, word)
		}

		if tmpQueryWords[len(tmpQueryWords)-1] != tmpQueryTags[len(tmpQueryWords)-1] {
			break
		}

		if !inListAlready(strings.ToLower(tmpQueryTags[queryWordCount]), tags) {
			tags = append(tags, strings.ToLower(tmpQueryTags[queryWordCount]))
		}

		counter++
	}

	c.JSON(200, gin.H{
		"tags":        tags,
		"suggestions": sr,
	})
}

func inListAlready(s string, list []string) bool {
	for _, s2 := range list {
		if s == s2 {
			return true
		}
	}
	return false
}

func isSeparator(sep string) bool {
	separatorsList := []string{
		"в", "с", "без",
	}
	for _, s := range separatorsList {
		if s == sep {
			return true
		}
	}
	return false
}
