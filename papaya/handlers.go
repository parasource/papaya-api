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
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/lightswitch/dutchman-backend/papaya/models"
	"github.com/sirupsen/logrus"
)

func (d *Papaya) registerRoutes(r *gin.Engine) {
	r.POST("/api/auth/register", d.HandleRegister)
	r.POST("/api/auth/login", d.HandleLogin)
	r.POST("/api/auth/refresh", d.AuthMiddleware, d.HandleRefresh)
	r.GET("/api/auth/user", d.AuthMiddleware, d.HandleUser)

	r.GET("/api/looks/:look", d.AuthMiddleware, d.HandleGetLook)
	r.GET("/api/looks/:look/item/:item", d.AuthMiddleware, d.HandleGetLookItem)
	r.PUT("/api/looks/:look/like", d.AuthMiddleware, d.HandleLikeLook)
	r.DELETE("/api/looks/:look/like", d.AuthMiddleware, d.HandleUnlikeLook)
	r.PUT("/api/looks/:look/dislike", d.AuthMiddleware, d.HandleDislikeLook)
	r.DELETE("/api/looks/:look/dislike", d.AuthMiddleware, d.HandleUndislikeLook)

	r.GET("/api/liked", d.AuthMiddleware, d.GetLikedLooks)

	// starting from page 0
	r.GET("/api/feed", d.AuthMiddleware, d.HandleFeed)

	r.GET("/api/test", d.HandleTest)

	// starting from page 0
	r.GET("/api/topics", d.AuthMiddleware, d.HandleGetTopics)
	r.GET("/api/topics/:topic", d.AuthMiddleware, d.HandleGetTopic)
	r.PUT("/api/topics/:topic/watch", d.AuthMiddleware, d.HandleWatchTopic)
	r.DELETE("/api/topics/:topic/unwatch", d.AuthMiddleware, d.HandleUnwatchTopic)

	r.POST("/api/collections/create", d.AuthMiddleware, d.HandleCreateCollection)
	r.GET("/api/collections/:collection", d.AuthMiddleware, d.HandleGetCollection)
	r.DELETE("/api/collections/:collection/delete", d.AuthMiddleware, d.HandleDeleteCollection)
	r.PUT("/api/collections/:collection/add/:look", d.AuthMiddleware, d.HandleCollectionAddLook)
	r.DELETE("/api/collections/:collection/remove/:look", d.AuthMiddleware, d.HandleCollectionRemoveLook)

	r.GET("/api/get-wardrobe-items", d.AuthMiddleware, d.HandleGetWardrobeItems)

	r.POST("/api/profile/set-wardrobe", d.AuthMiddleware, d.HandleProfileSetWardrobe)
	r.POST("/api/profile/set-mood", d.AuthMiddleware, d.HandleProfileSetMood)
	r.POST("/api/profile/update-settings", d.AuthMiddleware, d.HandleProfileUpdateSettings)
	r.GET("/api/profile/get-wardrobe", d.AuthMiddleware, d.HandleProfileGetWardrobe)
}

func (d *Papaya) HandleTest(c *gin.Context) {
	ids, err := d.adviser.RecommendForUser("1")
	if err != nil {
		logrus.Errorf("error inserting item: %v", err)
	}

	logrus.Info(ids)
}

// Root routes

func (d *Papaya) HandleGetWardrobeItems(c *gin.Context) {
	var items []*models.WardrobeCategory
	err := d.db.DB().Preload("Items").Find(&items).Error
	if err != nil {
		logrus.Errorf("error getting wardrobe items from db: %v", err)
		c.AbortWithStatus(500)
		return
	}

	var preview string
	for _, item := range items {
		d.db.DB().Raw("SELECT image FROM wardrobe_items WHERE wardrobe_category_id = ? ORDER BY id LIMIT 1", item.ID).Scan(&preview)
		item.Preview = preview
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
