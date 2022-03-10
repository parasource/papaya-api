package dutchman

import (
	"github.com/gin-gonic/gin"
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

	r.GET("/api/get-interests", d.AuthMiddleware, d.HandleGetAllInterests)
	r.GET("/api/profile/set-interests", d.AuthMiddleware, d.HandleProfileSetInterests)
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

func (d *Dutchman) HandleGetAllInterests(c *gin.Context) {
	interests := d.db.GetAllInterests()
	c.JSON(200, interests)
}

// Profile handlers

func (d *Dutchman) HandleProfileSetInterests(c *gin.Context) {

}

func (d *Dutchman) HandleProfileUpdateSettings(c *gin.Context) {

}
