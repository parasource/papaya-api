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

package papaya

import (
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/papaya/models"
	"github.com/lightswitch/dutchman-backend/papaya/util"
)

func (d *Papaya) GetUser(c *gin.Context) (*models.User, error) {
	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		return nil, err
	}

	claims, err := ParseToken(token)
	if err != nil {
		return nil, err
	}

	email := claims["email"].(string)
	user := d.db.GetUserByEmail(email)

	return user, nil
}
