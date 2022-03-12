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

package dutchman

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/sirupsen/logrus"
)

func (d *Dutchman) registerRoutes(r *gin.Engine) {
	r.POST("/api/auth/register", d.HandleRegister)
	r.POST("/api/auth/login", d.HandleLogin)
	r.POST("/api/auth/refresh", d.AuthMiddleware, d.HandleRefresh)
	r.GET("/api/auth/user", d.AuthMiddleware, d.HandleUser)

	r.GET("/api/looks/:look", d.AuthMiddleware, d.HandleGetLook)
	r.PUT("/api/looks/:look/favorites", d.AuthMiddleware, d.HandleAddLookToFavorites)
	r.DELETE("/api/looks/:look/favorites", d.AuthMiddleware, d.HandleRemoveLookFromFavorites)

	// starting from page 0
	r.GET("/api/feed/:page", d.AuthMiddleware, d.HandleFeed)
	r.GET("/api/selections", d.AuthMiddleware, d.HandleGetSelections)
	r.GET("/api/selections/:selection", d.AuthMiddleware, d.HandleGetSelection)
	r.PUT("/api/selections/:selection/favorites", d.AuthMiddleware, d.HandleAddSelectionToFavorites)
	r.DELETE("/api/selections/:selection/favorites", d.AuthMiddleware, d.HandleRemoveSelectionFromFavorites)

	r.GET("/api/get-wardrobe-items", d.AuthMiddleware, d.HandleGetWardrobeItems)

	r.POST("/api/profile/set-wardrobe", d.AuthMiddleware, d.HandleProfileSetWardrobe)
	r.POST("/api/profile/set-mood", d.AuthMiddleware, d.HandleProfileSetMood)
	r.POST("/api/profile/update-settings", d.AuthMiddleware, d.HandleProfileUpdateSettings)
}

// Root routes

func (d *Dutchman) HandleGetWardrobeItems(c *gin.Context) {
	var items []models.WardrobeItem
	err := d.db.DB().Find(&items).Error
	if err != nil {
		logrus.Errorf("error getting wardrobe items from db: %v", err)
		c.AbortWithStatus(500)
		return
	}
	c.JSON(200, items)
}

func ParseToken(accessToken string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}

		return []byte(accessSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalid
}

func getUserEmailFromToken(token string) (string, error) {
	claims, err := ParseToken(token)
	if err != nil {
		return "", err
	}

	id := claims["email"].(string)
	return id, nil
}
