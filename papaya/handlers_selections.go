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
	"github.com/sirupsen/logrus"
)

func (d *Dutchman) HandleGetSelections(c *gin.Context) {
	var result []models.Selection

	err := d.db.DB().Find(&result).Error
	if err != nil {
		logrus.Errorf("error getting all selections")
		c.AbortWithStatus(500)
	}

	c.JSON(200, result)
}

func (d *Dutchman) HandleGetSelection(c *gin.Context) {

}
