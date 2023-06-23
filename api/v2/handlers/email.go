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
	"net/http"
	"strings"
)

type EmailSubscribeRequest struct {
	Email string `json:"email"`
}

func HandleEmailSubscribe(c *gin.Context) {
	var r EmailSubscribeRequest
	err := c.ShouldBindJSON(&r)
	if err != nil || strings.TrimSpace(r.Email) == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var exists bool
	database.DB().Raw("select exists(select 1 from email_subscriptions where email = ?)", r.Email).Scan(&exists)
	if exists {
		c.JSON(http.StatusNoContent, gin.H{})
	}

	sub := models.EmailSubscription{
		Email:    r.Email,
		IsActive: true,
	}
	err = database.DB().Create(&sub).Error
	if err != nil {
		log.Error().Err(err).Msg("error creating email subscription")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusNoContent, gin.H{})
}
