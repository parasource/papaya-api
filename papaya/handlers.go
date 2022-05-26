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

package papaya

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/lightswitch/dutchman-backend/papaya/models"
	"github.com/sirupsen/logrus"
)

func (p *Papaya) registerRoutes(r *gin.Engine) {
	r.POST("/api/auth/register", p.HandleRegister)
	r.POST("/api/auth/login", p.HandleLogin)
	r.POST("/api/auth/refresh", p.AuthMiddleware, p.HandleRefresh)
	r.GET("/api/auth/user", p.AuthMiddleware, p.HandleUser)

	r.GET("/api/looks/:look", p.AuthMiddleware, p.HandleGetLook)
	r.GET("/api/looks/:look/item/:item", p.AuthMiddleware, p.HandleGetLookItem)
	r.PUT("/api/looks/:look/like", p.AuthMiddleware, p.HandleLikeLook)
	r.DELETE("/api/looks/:look/like", p.AuthMiddleware, p.HandleUnlikeLook)
	r.PUT("/api/looks/:look/dislike", p.AuthMiddleware, p.HandleDislikeLook)
	r.DELETE("/api/looks/:look/dislike", p.AuthMiddleware, p.HandleUndislikeLook)

	r.GET("/api/liked", p.AuthMiddleware, p.GetLikedLooks)

	// starting from page 0
	r.GET("/api/feed", p.AuthMiddleware, p.HandleFeed)
	r.GET("/api/feed/:tag", p.AuthMiddleware, p.HandleFeedByTag)

	r.GET("/api/test", p.HandleTest)

	r.GET("/api/search", p.HandleSearch)

	// starting from page 0
	r.GET("/api/topics/recommended", p.AuthMiddleware, p.HandleGetRecommendedTopics)
	r.GET("/api/topics/popular", p.AuthMiddleware, p.HandleGetPopularTopics)
	r.GET("/api/topics/saved", p.AuthMiddleware, p.HandleGetSavedTopics)
	r.GET("/api/topics/:topic", p.AuthMiddleware, p.HandleGetTopic)
	r.PUT("/api/topics/:topic/save", p.AuthMiddleware, p.HandleSaveTopic)
	r.DELETE("/api/topics/:topic/unsave", p.AuthMiddleware, p.HandleUnsaveTopic)

	r.POST("/api/collections", p.AuthMiddleware, p.HandleGetCollections)
	r.POST("/api/collections/create", p.AuthMiddleware, p.HandleCreateCollection)
	r.GET("/api/collections/:collection", p.AuthMiddleware, p.HandleGetCollection)
	r.DELETE("/api/collections/:collection/delete", p.AuthMiddleware, p.HandleDeleteCollection)
	r.PUT("/api/collections/:collection/add/:look", p.AuthMiddleware, p.HandleCollectionAddLook)
	r.DELETE("/api/collections/:collection/remove/:look", p.AuthMiddleware, p.HandleCollectionRemoveLook)

	r.GET("/api/get-wardrobe-items", p.AuthMiddleware, p.HandleGetWardrobeItems)

	r.POST("/api/profile/set-wardrobe", p.AuthMiddleware, p.HandleProfileSetWardrobe)
	r.POST("/api/profile/set-mood", p.AuthMiddleware, p.HandleProfileSetMood)
	r.POST("/api/profile/update-settings", p.AuthMiddleware, p.HandleProfileUpdateSettings)
	r.GET("/api/profile/get-wardrobe", p.AuthMiddleware, p.HandleProfileGetWardrobe)
}

func (p *Papaya) HandleTest(c *gin.Context) {
	ids, err := p.adviser.RecommendForUser("1")
	if err != nil {
		logrus.Errorf("error inserting item: %v", err)
	}

	logrus.Info(ids)
}

// Root routes

func (p *Papaya) HandleGetWardrobeItems(c *gin.Context) {
	var items []*models.WardrobeCategory
	err := p.db.DB().Preload("Items").Find(&items).Error
	if err != nil {
		logrus.Errorf("error getting wardrobe items from db: %v", err)
		c.AbortWithStatus(500)
		return
	}

	var preview string
	for _, item := range items {
		p.db.DB().Raw("SELECT image FROM wardrobe_items WHERE wardrobe_category_id = ? ORDER BY id LIMIT 1", item.ID).Scan(&preview)
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
