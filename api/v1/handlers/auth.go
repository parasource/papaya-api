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
	"encoding/json"
	"fmt"
	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/parasource/papaya-api/api/v1/requests"
	"github.com/parasource/papaya-api/pkg/database"
	"github.com/parasource/papaya-api/pkg/database/models"
	"github.com/parasource/papaya-api/pkg/util"
	"github.com/sirupsen/logrus"
	"io/ioutil"
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
	//acf := NewAuthCodeFlowUser(oauth.UserParams{
	//	ClientID:    123456,
	//	RedirectURI: "https://example.com/callback",
	//	Scope:       oauth.ScopeUserPhotos + oauth.ScopeUserDocs,
	//}, clientSecret)

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

func HandleGoogleLoginOrRegister(c *gin.Context) {
	var r requests.GoogleUserInput
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding google request: %v", err)
		c.AbortWithStatus(500)
		return
	}

	endpoint := "https://www.googleapis.com/userinfo/v2/me"

	client := &http.Client{}
	req, _ := http.NewRequest("GET", endpoint, nil)
	header := "Bearer " + r.AccessToken
	req.Header.Set("Authorization", header)
	res, googleErr := client.Do(req)
	if googleErr != nil {
		logrus.Errorf("error getting user information from google while signing up: %v", err)
		c.JSON(500, gin.H{
			"success": false,
		})
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logrus.Errorf("error reading google response body: %v", err)
		c.JSON(500, gin.H{
			"success": false,
		})
		return
	}

	var googleRes requests.GoogleUserRes
	json.Unmarshal(body, &googleRes)

	if googleRes.Email != "" {
		var user models.User
		err = database.DB().Model(&user).Where("email = ?", googleRes.Email).First(&user).Error
		if err != nil {
			logrus.Errorf("error getting user by email while signing up: %v", err)
		}

		// User is not found, so we'll sign him up
		if user.ID == 0 {
			user := models.NewUser(googleRes.Email, googleRes.Name, "")
			user.Sex = googleRes.Gender
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
		} else {
			token, err := util.GenerateToken(&user)
			if err != nil {
				logrus.Errorf("error generating token: %v", err)
				c.JSON(500, gin.H{
					"success": false,
				})
				return
			}
			refreshToken, err := util.GenerateRefreshToken(user.ID)
			if err != nil {
				logrus.Errorf("error generating refresh token: %v", err)
				c.JSON(500, gin.H{
					"success": false,
				})
				return
			}

			c.JSON(200, gin.H{
				"success":       true,
				"token":         token,
				"refresh_token": refreshToken,
			})
		}
	}
}

func HandleAppleLoginOrRegister(c *gin.Context) {
	var r requests.AppleUserInput
	err := c.BindJSON(&r)
	if err != nil {
		logrus.Errorf("error binding google request: %v", err)
		c.AbortWithStatus(500)
		return
	}

	res, httpErr := http.Get("https://appleid.apple.com/auth/keys")
	if httpErr != nil {
		logrus.Errorf("error sending request to apple auth service: %v", err)
		c.JSON(500, gin.H{
			"success": false,
		})
		return
	}

	defer res.Body.Close()

	body, bodyErr := ioutil.ReadAll(res.Body)
	if bodyErr != nil {
		logrus.Errorf("error reading response body: %v", err)
		c.JSON(500, gin.H{
			"success": false,
		})
		return
	}

	jwks, err := keyfunc.NewJSON(body)
	if err != nil {
		logrus.Errorf("keyfunc error: %v", err)
		c.JSON(500, gin.H{
			"success": false,
		})
		return
	}
	token, err := jwt.Parse(r.IdentityToken, jwks.Keyfunc)
	if err != nil {
		logrus.Errorf("error parsing apple token: %v", err)
		c.JSON(500, gin.H{
			"success": false,
		})
		return
	}

	if !token.Valid {
		logrus.Errorf("apple invalid token: %v", err)
		c.JSON(500, gin.H{
			"success": false,
		})
		return
	}

	email := fmt.Sprint(token.Claims.(jwt.MapClaims)["email"])
	name := fmt.Sprint(token.Claims.(jwt.MapClaims)["full_name"])
	logrus.Info("response from apple: %v", token.Claims.(jwt.MapClaims))

	if email != "" {
		var user models.User
		err = database.DB().Model(&user).Where("email = ?", email).First(&user).Error
		if err != nil {
			logrus.Errorf("error getting user by email while signing up: %v", err)
		}

		// User is not found, so we'll sign him up
		if user.ID == 0 {
			user := models.NewUser(email, name, "")
			user.Sex = "male"
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
		} else {
			token, err := util.GenerateToken(&user)
			if err != nil {
				logrus.Errorf("error generating token: %v", err)
				c.JSON(500, gin.H{
					"success": false,
				})
				return
			}
			refreshToken, err := util.GenerateRefreshToken(user.ID)
			if err != nil {
				logrus.Errorf("error generating refresh token: %v", err)
				c.JSON(500, gin.H{
					"success": false,
				})
				return
			}

			c.JSON(200, gin.H{
				"success":       true,
				"token":         token,
				"refresh_token": refreshToken,
			})
		}
	}
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
