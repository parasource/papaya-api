package dutchman

import (
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
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

	r.GET("/api/profile/set-interests", d.AuthMiddleware, d.HandleProfileSetInterests)
	r.GET("/api/profile/update-settings", d.AuthMiddleware, d.HandleProfileUpdateSettings)
}

func (d *Dutchman) HandleRegister(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")
	password := c.PostForm("password")

	if d.db.CheckIfUserExists(email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Пользователь с таким адресом эл.почты уже существует",
		})
		return
	}

	user := models.NewUser(email, name, password)
	err := d.db.StoreUser(user)
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
	email := c.PostForm("email")
	password := c.PostForm("password")

	user := d.db.GetUserByEmail(email)
	if user == nil {
		c.JSON(403, gin.H{
			"success": false,
			"message": "Пользователь не найден",
		})
		return
	}
	// Wrong password
	if !user.CheckPasswordHash(password) {
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

// Profile handlers

func (d *Dutchman) HandleProfileSetInterests(c *gin.Context) {

}

func (d *Dutchman) HandleProfileUpdateSettings(c *gin.Context) {

}
