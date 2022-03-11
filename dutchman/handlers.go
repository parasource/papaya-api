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
	"github.com/lightswitch/dutchman-backend/dutchman/requests"
	"github.com/lightswitch/dutchman-backend/dutchman/util"
	"github.com/sirupsen/logrus"
	"net/http"
)

func (d *Dutchman) registerRoutes(r *gin.Engine) {
	r.POST("/api/auth/register", d.HandleRegister)
	r.POST("/api/auth/login", d.HandleLogin)
	r.POST("/api/auth/refresh", d.AuthMiddleware, d.HandleRefresh)
	r.GET("/api/auth/user", d.AuthMiddleware, d.HandleUser)

	r.GET("/api/feed", d.AuthMiddleware, d.HandleFeed)

	r.GET("/api/get-wardrobe-items", d.AuthMiddleware, d.HandleGetWardrobeItems)

	r.POST("/api/profile/set-wardrobe", d.AuthMiddleware, d.HandleProfileSetWardrobe)
	r.POST("/api/profile/set-mood", d.AuthMiddleware, d.HandleProfileSetMood)

	r.GET("/api/profile/update-settings", d.AuthMiddleware, d.HandleProfileUpdateSettings)
}

func (d *Dutchman) HandleRegister(c *gin.Context) {
	var r requests.RegisterRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding register request: %v", err)
		return
	}
	logrus.Infof("req %v", r)

	if d.db.CheckIfUserExists(r.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Пользователь с таким адресом эл.почты уже существует",
		})
		return
	}

	user := models.NewUser(r.Email, r.Name, r.Password)
	err = d.db.StoreUser(user)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	token, err := GenerateToken(user)
	if err != nil {
		logrus.Errorf("error generating token: %v", err)
		c.AbortWithStatus(500)
		return
	}
	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		logrus.Errorf("error generating refresh token: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"success":       true,
		"token":         token,
		"refresh_token": refreshToken,
	})
}

func (d *Dutchman) HandleLogin(c *gin.Context) {
	var r requests.LoginRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding login request: %v", err)
		c.AbortWithStatus(500)
		return
	}
	logrus.Debug(r)

	user := d.db.GetUserByEmail(r.Email)
	if user == nil {
		c.JSON(403, gin.H{
			"success": false,
			"message": "Пользователь не найден",
		})
		return
	}
	// Wrong password
	if !user.CheckPasswordHash(r.Password) {
		c.JSON(403, gin.H{
			"success": false,
			"message": "Неверный пароль",
		})
		return
	}
	token, err := GenerateToken(user)
	if err != nil {
		logrus.Errorf("error generating token: %v", err)
		c.AbortWithStatus(500)
		return
	}
	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		logrus.Errorf("error generating refresh token: %v", err)
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"success":       true,
		"token":         token,
		"refresh_token": refreshToken,
	})
}

func (d *Dutchman) HandleRefresh(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")

	if refreshToken == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	claims, err := ParseRefreshToken(refreshToken)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := claims["id"].(string)
	user := d.db.GetUser(id)
	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Now we generate a new token for user
	token, err := GenerateToken(user)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	refreshTokenNew, err := GenerateRefreshToken(user.ID)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}

	c.JSON(200, gin.H{
		"success":       true,
		"token":         token,
		"refresh_token": refreshTokenNew,
	})
}

func (d *Dutchman) HandleUser(c *gin.Context) {
	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, err := ParseToken(token)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	id := claims["id"].(string)
	user := d.db.GetUser(id)
	if user == nil {
		logrus.Errorf("user with claims not found: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(200, user)
}

func (d *Dutchman) HandleFeed(c *gin.Context) {

}

// All routes

func (d *Dutchman) HandleGetWardrobeItems(c *gin.Context) {
	interests := d.db.GetWardrobeItems()
	c.JSON(200, interests)
}

// Profile handlers

func (d *Dutchman) HandleProfileSetWardrobe(c *gin.Context) {
	var r requests.SetWardrobeRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile wardrobe request: %v", err)
		return
	}

	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId, err := getUserIdFromToken(token)
	if err != nil {
		c.AbortWithStatus(500)
	}
	err = d.db.SetUserWardrobe(userId, r.Wardrobe)
	if err != nil {
		c.AbortWithStatus(500)
		logrus.Errorf("error setting profile mood: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleProfileSetMood(c *gin.Context) {
	var r requests.SetMoodRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding set profile mood request: %v", err)
		return
	}

	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId, err := getUserIdFromToken(token)
	if err != nil {
		c.AbortWithStatus(500)
	}
	err = d.db.SetUserMood(userId, r.Mood)
	if err != nil {
		c.AbortWithStatus(500)
		logrus.Errorf("error setting profile mood: %v", err)
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}

func (d *Dutchman) HandleProfileUpdateSettings(c *gin.Context) {

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

func getUserIdFromToken(token string) (string, error) {
	claims, err := ParseToken(token)
	if err != nil {
		return "", err
	}

	id := claims["id"].(string)
	return id, nil
}
