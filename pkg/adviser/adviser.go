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
	"github.com/go-redis/redis/v9"
	"github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/parasource/papaya-api/pkg/gorse"
	"github.com/sirupsen/logrus"
	"math/rand"
	"strconv"
	"time"
)

var instance *Adviser

// 1 - user id
// 2 - user id
// 3 - user sex
// 4 - limit
// 5 - offset
const feedWardrobeRecommendationTemplate = `select looks.* from looks
    join look_items li on looks.id = li.look_id
    right join users_wardrobe uw on li.wardrobe_item_id = uw.wardrobe_item_id
               WHERE uw.user_id = ?
                 AND looks.id NOT IN (SELECT saved_looks.look_id FROM saved_looks WHERE saved_looks.user_id = ?)
                 AND looks.sex = ?
                 AND looks.deleted_at IS NULL
               GROUP BY looks.id ORDER BY looks.id DESC LIMIT ? OFFSET ?;`

type Adviser struct {
	cache *redis.Client
}

func Get() *Adviser {
	if instance == nil {
		instance = &Adviser{}
	}
	return instance
}

func (a *Adviser) Feed(user *models.User, limit int, offset int) ([]*models.Look, error) {
	var looks []*models.Look

	// So first we grab half of page items from recommendations

	err := database.DB().Debug().Raw(feedWardrobeRecommendationTemplate, user.ID, user.ID, user.Sex, limit/2, offset/2).Scan(&looks).Error
	if err != nil {
		return nil, err
	}
	if len(looks) == 0 {
		return looks, nil
	}

	// Then, if there are still wardrobe recommendations, we complete them with gorse recommendations
	var looks1 []*models.Look
	items, err := gorse.RecommendForUserAndCategory(strconv.Itoa(int(user.ID)), user.Sex, limit/2, offset/2)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		logrus.Debug("did not recommend anything")

		err = database.DB().Debug().Raw("SELECT * FROM looks WHERE sex = ? ORDER BY random() LIMIT ? OFFSET ?", user.Sex, limit/2, offset/2).Scan(&looks1).Error
	} else {
		err = database.DB().Debug().Where("slug IN ?", items).Find(&looks1).Error
		if err != nil {
			logrus.Errorf("error finding looks by recommendation: %v", err)
		}
	}
	looks = append(looks, looks1...)

	// Random sorting for entropy
	rand.Seed(time.Now().UnixNano())
	for i := len(looks) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := rand.Intn(i + 1)
		looks[i], looks[j] = looks[j], looks[i]
	}

	return looks, nil
}
