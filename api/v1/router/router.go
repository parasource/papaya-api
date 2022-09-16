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

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/parasource/papaya-api/api/v1/handlers"
	"github.com/parasource/papaya-api/api/v1/middleware"
)

func Initialize() *gin.Engine {
	r := gin.Default()

	/// Authentication & Authorization
	r.POST("/api/auth/register", handlers.HandleRegister)
	r.POST("/api/auth/login", handlers.HandleLogin)
	r.POST("/api/auth/refresh", middleware.AuthMiddleware, handlers.HandleRefresh)
	r.GET("/api/auth/user", middleware.AuthMiddleware, handlers.HandleUser)

	/// Search
	r.GET("/api/search", middleware.AuthMiddleware, handlers.HandleSearch)
	r.GET("/api/search/popular", middleware.AuthMiddleware, handlers.HandleSearchPopular)
	r.GET("/api/search/history", middleware.AuthMiddleware, handlers.HandleSearchHistory)
	r.GET("/api/search/suggest", middleware.AuthMiddleware, handlers.HandleSearchSuggestions)

	/// Topics
	r.GET("/api/topics/recommended", middleware.AuthMiddleware, handlers.HandleGetRecommendedTopics)
	r.GET("/api/topics/popular", middleware.AuthMiddleware, handlers.HandleGetPopularTopics)
	r.GET("/api/topics/saved", middleware.AuthMiddleware, handlers.HandleGetSavedTopics)
	r.GET("/api/topics/:topic", middleware.AuthMiddleware, handlers.HandleGetTopic)
	r.PUT("/api/topics/:topic/save", middleware.AuthMiddleware, handlers.HandleSaveTopic)
	r.DELETE("/api/topics/:topic/unsave", middleware.AuthMiddleware, handlers.HandleUnsaveTopic)

	/// Saved

	r.GET("/api/saved", middleware.AuthMiddleware, handlers.HandleSaved)
	r.POST("/api/saved/:look", middleware.AuthMiddleware, handlers.HandleSavedAdd)
	r.DELETE("/api/saved/:look", middleware.AuthMiddleware, handlers.HandleSavedRemove)

	/// Collections
	r.GET("/api/collections", middleware.AuthMiddleware, handlers.HandleGetCollections)
	r.POST("/api/collections/create", middleware.AuthMiddleware, handlers.HandleCreateCollection)
	r.DELETE("/api/collections/:collection/delete", middleware.AuthMiddleware, handlers.HandleDeleteCollection)
	r.PUT("/api/collections/:collection/add/:look", middleware.AuthMiddleware, handlers.HandleCollectionAddLook)
	r.DELETE("/api/collections/:collection/remove/:look", middleware.AuthMiddleware, handlers.HandleCollectionRemoveLook)
	r.GET("/api/collections/:collection", middleware.AuthMiddleware, handlers.HandleGetCollection)

	/// Feed and looks
	r.GET("/api/looks/:look", middleware.AuthMiddleware, handlers.HandleGetLook)
	r.GET("/api/looks/:look/item/:item", middleware.AuthMiddleware, handlers.HandleGetLookItem)
	r.PUT("/api/looks/:look/like", middleware.AuthMiddleware, handlers.HandleLikeLook)
	r.DELETE("/api/looks/:look/like", middleware.AuthMiddleware, handlers.HandleUnlikeLook)
	r.PUT("/api/looks/:look/dislike", middleware.AuthMiddleware, handlers.HandleDislikeLook)
	r.DELETE("/api/looks/:look/dislike", middleware.AuthMiddleware, handlers.HandleUndislikeLook)
	r.GET("/api/liked", middleware.AuthMiddleware, handlers.GetLikedLooks)
	r.GET("/api/feed", middleware.AuthMiddleware, handlers.HandleFeed)
	r.GET("/api/feed/:category", middleware.AuthMiddleware, handlers.HandleFeedByCategory)
	r.GET("/api/get-wardrobe-items", middleware.AuthMiddleware, handlers.HandleGetWardrobeItems)

	/// Profile
	r.POST("/api/profile/set-wardrobe", middleware.AuthMiddleware, handlers.HandleProfileSetWardrobe)
	r.POST("/api/profile/set-mood", middleware.AuthMiddleware, handlers.HandleProfileSetMood)
	r.POST("/api/profile/update-settings", middleware.AuthMiddleware, handlers.HandleProfileUpdateSettings)
	r.GET("/api/profile/get-wardrobe", middleware.AuthMiddleware, handlers.HandleProfileGetWardrobe)
	return r
}
