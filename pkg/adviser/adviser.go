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

package adviser

import (
	"github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/parasource/papaya-api/pkg/gorse"
	"math/rand"
	"strconv"
	"time"
)

var instance *Adviser

const feedWardrobeRecommendationTemplate = `select looks.* from looks
    join look_items li on looks.id = li.look_id
    right join users_wardrobe uw on li.wardrobe_item_id = uw.wardrobe_item_id
               WHERE uw.user_id = ?
                 AND looks.id NOT IN (SELECT saved_looks.look_id FROM saved_looks WHERE saved_looks.user_id = ?)
                 AND looks.sex = ?
                 AND looks.deleted_at IS NULL
               	 AND looks.slug NOT IN ?
               GROUP BY looks.id ORDER BY looks.id DESC LIMIT ? OFFSET ?;`

const feedWardrobeRecommendationFallbackTemplate = `select looks.* from looks
    join look_items li on looks.id = li.look_id
    right join users_wardrobe uw on li.wardrobe_item_id = uw.wardrobe_item_id
               WHERE uw.user_id = ?
                 AND looks.id NOT IN (SELECT saved_looks.look_id FROM saved_looks WHERE saved_looks.user_id = ?)
                 AND looks.sex = ?
                 AND looks.deleted_at IS NULL
               GROUP BY looks.id ORDER BY looks.id DESC LIMIT ? OFFSET ?;`

type Adviser struct {
	cache *Cache
}

func Get() *Adviser {
	if instance == nil {
		instance = &Adviser{}
	}
	return instance
}

func (a *Adviser) Feed(user *models.User, page int) ([]*models.Look, error) {
	var looks []*models.Look

	// So first we grab major part of page items from gorse
	slugs, err := gorse.RecommendForUserAndCategory(strconv.Itoa(int(user.ID)), user.Sex, 15, 15*page)
	if err != nil {
		return nil, err
	}
	err = database.DB().Where("slug in ?", slugs).Find(&looks).Error
	if err != nil {
		return nil, err
	}

	// Then we need looks, which contain at least one item from user's wardrobe
	var wardrobeLooks []*models.Look

	// This is the main scenario, but we need a fallback when there are no recommendation from gorse
	if len(slugs) > 0 {
		err = database.DB().Debug().Raw(feedWardrobeRecommendationTemplate, user.ID, user.ID, user.Sex, slugs, 5, 5*page).Scan(&wardrobeLooks).Error
		if err != nil {
			return nil, err
		}
	} else {
		err = database.DB().Debug().Raw(feedWardrobeRecommendationFallbackTemplate, user.ID, user.ID, user.Sex, 20, 20*page).Scan(&wardrobeLooks).Error
		if err != nil {
			return nil, err
		}
	}
	for _, look := range wardrobeLooks {
		look.IsFromWardrobe = true
	}

	looks = append(looks, wardrobeLooks...)

	// Random sorting for entropy
	rand.Seed(time.Now().UnixNano())
	for i := len(looks) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.Intn(i + 1)
		looks[i], looks[j] = looks[j], looks[i]
	}

	return looks, nil
}
