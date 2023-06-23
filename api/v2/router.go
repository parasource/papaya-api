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

package v2

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/parasource/papaya-api/api/v2/handlers"
	"github.com/parasource/papaya-api/api/v2/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

func Routes(r *gin.Engine) {
	apiV2 := r.Group("/api/v2")

	apiV2.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"*"},
	}))

	apiV2.POST("/frontend/error-logs", func(c *gin.Context) {
		type FrontendError struct {
			Error   string `json:"error"`
			IsFatal bool   `json:"isFatal"`
		}
		var err FrontendError
		if jsonErr := c.ShouldBindJSON(&err); jsonErr != nil {
			log.Error().Err(jsonErr).Msg("invalid format")
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		log.Error().Str("error", err.Error).Bool("is_fatal", err.IsFatal).Msg("received fronted error")
	})

	/// Authentication & Authorization
	apiV2.POST("/auth/register", handlers.HandleRegister)
	apiV2.POST("/auth/login", handlers.HandleLogin)
	apiV2.POST("/auth/login/google", handlers.HandleGoogleLoginOrRegister)
	apiV2.POST("/auth/login/apple", handlers.HandleAppleLoginOrRegister)
	apiV2.POST("/auth/login/vk", handlers.HandleLogin)
	apiV2.POST("/auth/refresh", handlers.HandleRefresh)
	apiV2.GET("/auth/user", middleware.AuthMiddleware, handlers.HandleUser)

	/// Search
	apiV2.GET("/search", middleware.AuthMiddleware, handlers.HandleSearch)
	apiV2.GET("/search/suggestions", middleware.AuthMiddleware, handlers.HandleSearchSuggestions)
	apiV2.POST("/search/clear-history", middleware.AuthMiddleware, handlers.HandleSearchClearHistory)
	apiV2.GET("/search/autofill", middleware.AuthMiddleware, handlers.HandleSearchAutofill)

	/// Topics
	apiV2.GET("/topics/saved", middleware.AuthMiddleware, handlers.HandleGetSavedTopics)
	apiV2.GET("/topics/:topic", middleware.AuthMiddleware, handlers.HandleGetTopic)
	apiV2.PUT("/topics/:topic/save", middleware.AuthMiddleware, handlers.HandleSaveTopic)
	apiV2.DELETE("/topics/:topic/unsave", middleware.AuthMiddleware, handlers.HandleUnsaveTopic)

	/// Saved
	apiV2.GET("/saved", middleware.AuthMiddleware, handlers.HandleSaved)
	apiV2.POST("/saved/:look", middleware.AuthMiddleware, handlers.HandleSavedAdd)
	apiV2.DELETE("/saved/:look", middleware.AuthMiddleware, handlers.HandleSavedRemove)

	/// Feed and looks
	apiV2.GET("/looks/:look", middleware.AuthMiddleware, handlers.HandleGetLook)
	apiV2.GET("/looks/:look/item/:item", middleware.AuthMiddleware, handlers.HandleGetLookItem)
	apiV2.PUT("/looks/:look/like", middleware.AuthMiddleware, handlers.HandleLikeLook)
	apiV2.DELETE("/looks/:look/like", middleware.AuthMiddleware, handlers.HandleUnlikeLook)
	apiV2.PUT("/looks/:look/dislike", middleware.AuthMiddleware, handlers.HandleDislikeLook)
	apiV2.DELETE("/looks/:look/dislike", middleware.AuthMiddleware, handlers.HandleUndislikeLook)
	apiV2.GET("/liked", middleware.AuthMiddleware, handlers.GetLikedLooks)
	apiV2.GET("/feed", middleware.AuthMiddleware, handlers.HandleFeed)
	apiV2.GET("/feed/:category", middleware.AuthMiddleware, handlers.HandleFeedByCategory)

	// Articles
	// I'll make it open because we have an articles site,
	// and we don't need to close it
	apiV2.GET("/articles", handlers.HandleGetArticles)
	apiV2.GET("/articles/:slug", handlers.HandleGetArticle)
	apiV2.GET("/articles/search", handlers.HandleSearchArticles)

	/// Wardrobe
	apiV2.GET("/wardrobe", middleware.AuthMiddleware, handlers.HandleGetWardrobeCategories)
	apiV2.GET("/wardrobe/:category", middleware.AuthMiddleware, handlers.HandleGetWardrobeItems)

	// Email Subscriptions
	apiV2.POST("/email/subscribe", handlers.HandleEmailSubscribe)

	/// Profile
	apiV2.POST("/profile/set-wardrobe", middleware.AuthMiddleware, handlers.HandleProfileSetWardrobe)
	apiV2.POST("/profile/set-mood", middleware.AuthMiddleware, handlers.HandleProfileSetMood)
	apiV2.POST("/profile/update-settings", middleware.AuthMiddleware, handlers.HandleProfileUpdateSettings)
	apiV2.GET("/profile/get-wardrobe", middleware.AuthMiddleware, handlers.HandleProfileGetWardrobe)
	apiV2.POST("/profile/set-apns-token", middleware.AuthMiddleware, handlers.HandleSetAPNSToken)
}
