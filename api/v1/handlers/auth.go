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

package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/papaya-api/api/v1/requests"
	"github.com/lightswitch/papaya-api/pkg/db"
	"github.com/lightswitch/papaya-api/pkg/db/models"
	"github.com/lightswitch/papaya-api/pkg/util"
	"github.com/sirupsen/logrus"
	"net/http"
)

func HandleRegister(c *gin.Context) {
	var r requests.RegisterRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding register request: %v", err)
		return
	}
	logrus.Infof("req %v", r)

	if database.GetUserByEmail(r.Email) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Пользователь с таким адресом эл.почты уже существует",
		})
		return
	}

	user := models.NewUser(r.Email, r.Name, r.Password)
	user.Sex = r.Sex
	database.CreateUser(user)

	token, err := util.GenerateToken(user)
	if err != nil {
		logrus.Errorf("error generating token: %v", err)
		c.AbortWithStatus(500)
		return
	}
	refreshToken, err := util.GenerateRefreshToken(user.ID)
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

func HandleAuthVK(c *gin.Context) {
	//acf := oauth.NewAuthCodeFlowUser(oauth.UserParams{
	//	ClientID: 123123,
	//	RedirectURI: "https://papaya.io",
	//	Scope: oauth.ScopeUserPhotos + oauth.ScopeUserEmail,
	//}, "")
	//
	//
	//c.JSON(200, )
}

func HandleLogin(c *gin.Context) {
	var r requests.LoginRequest
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding login request: %v", err)
		c.AbortWithStatus(500)
		return
	}
	logrus.Debug(r)

	user := database.GetUserByEmail(r.Email)
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
	token, err := util.GenerateToken(user)
	if err != nil {
		logrus.Errorf("error generating token: %v", err)
		c.AbortWithStatus(500)
		return
	}
	refreshToken, err := util.GenerateRefreshToken(user.ID)
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

func HandleRefresh(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")

	if refreshToken == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	claims, err := util.ParseRefreshToken(refreshToken)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id := claims["id"].(uint)
	user := database.GetUser(id)
	if user == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Now we generate a new token for user
	token, err := util.GenerateToken(user)
	if err != nil {
		c.AbortWithStatus(500)
		return
	}
	refreshTokenNew, err := util.GenerateRefreshToken(user.ID)
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

func HandleUser(c *gin.Context) {
	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	claims, err := util.ParseToken(token)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	email := claims["email"].(string)
	user := database.GetUserByEmail(email)
	if user == nil {
		logrus.Errorf("user with claims not found: %v", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.JSON(200, user)
}

///////////////////
/// Helper methods

func GetUser(c *gin.Context) (*models.User, error) {
	token, err := util.ExtractToken(c.GetHeader("Authorization"))
	if err != nil {
		return nil, err
	}

	claims, err := util.ParseToken(token)
	if err != nil {
		return nil, err
	}

	email := claims["email"].(string)
	user := database.GetUserByEmail(email)

	return user, nil
}
