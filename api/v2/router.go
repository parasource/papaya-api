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
	"github.com/gin-gonic/gin"
	"github.com/parasource/papaya-api/api/v2/handlers"
	"github.com/parasource/papaya-api/api/v2/middleware"
)

func Routes(r *gin.Engine) {
	apiV1 := r.Group("/api/v2")

	/// Authentication & Authorization
	apiV1.POST("/auth/register", handlers.HandleRegister)
	apiV1.POST("/auth/login", handlers.HandleLogin)
	apiV1.POST("/auth/login/google", handlers.HandleGoogleLoginOrRegister)
	apiV1.POST("/auth/login/apple", handlers.HandleAppleLoginOrRegister)
	apiV1.POST("/auth/login/vk", handlers.HandleLogin)
	apiV1.POST("/auth/refresh", handlers.HandleRefresh)
	apiV1.GET("/auth/user", middleware.AuthMiddleware, handlers.HandleUser)

	/// Search
	apiV1.GET("/search", middleware.AuthMiddleware, handlers.HandleSearch)
	apiV1.GET("/search/suggestions", middleware.AuthMiddleware, handlers.HandleSearchSuggestions)
	apiV1.POST("/search/clear-history", middleware.AuthMiddleware, handlers.HandleSearchClearHistory)
	apiV1.GET("/search/autofill", middleware.AuthMiddleware, handlers.HandleSearchAutofill)

	/// Topics
	apiV1.GET("/topics/saved", middleware.AuthMiddleware, handlers.HandleGetSavedTopics)
	apiV1.GET("/topics/:topic", middleware.AuthMiddleware, handlers.HandleGetTopic)
	apiV1.PUT("/topics/:topic/save", middleware.AuthMiddleware, handlers.HandleSaveTopic)
	apiV1.DELETE("/topics/:topic/unsave", middleware.AuthMiddleware, handlers.HandleUnsaveTopic)

	/// Saved
	apiV1.GET("/saved", middleware.AuthMiddleware, handlers.HandleSaved)
	apiV1.POST("/saved/:look", middleware.AuthMiddleware, handlers.HandleSavedAdd)
	apiV1.DELETE("/saved/:look", middleware.AuthMiddleware, handlers.HandleSavedRemove)

	/// Feed and looks
	apiV1.GET("/looks/:look", middleware.AuthMiddleware, handlers.HandleGetLook)
	apiV1.GET("/looks/:look/item/:item", middleware.AuthMiddleware, handlers.HandleGetLookItem)
	apiV1.PUT("/looks/:look/like", middleware.AuthMiddleware, handlers.HandleLikeLook)
	apiV1.DELETE("/looks/:look/like", middleware.AuthMiddleware, handlers.HandleUnlikeLook)
	apiV1.PUT("/looks/:look/dislike", middleware.AuthMiddleware, handlers.HandleDislikeLook)
	apiV1.DELETE("/looks/:look/dislike", middleware.AuthMiddleware, handlers.HandleUndislikeLook)
	apiV1.GET("/liked", middleware.AuthMiddleware, handlers.GetLikedLooks)
	apiV1.GET("/feed", middleware.AuthMiddleware, handlers.HandleFeed)
	apiV1.GET("/feed/:category", middleware.AuthMiddleware, handlers.HandleFeedByCategory)

	/// Wardrobe
	apiV1.GET("/wardrobe", middleware.AuthMiddleware, handlers.HandleGetWardrobeCategories)
	apiV1.GET("/wardrobe/:category", middleware.AuthMiddleware, handlers.HandleGetWardrobeItems)

	/// Profile
	apiV1.POST("/profile/set-wardrobe", middleware.AuthMiddleware, handlers.HandleProfileSetWardrobe)
	apiV1.POST("/profile/set-mood", middleware.AuthMiddleware, handlers.HandleProfileSetMood)
	apiV1.POST("/profile/update-settings", middleware.AuthMiddleware, handlers.HandleProfileUpdateSettings)
	apiV1.GET("/profile/get-wardrobe", middleware.AuthMiddleware, handlers.HandleProfileGetWardrobe)
	apiV1.POST("/profile/set-apns-token", middleware.AuthMiddleware, handlers.HandleSetAPNSToken)
}
