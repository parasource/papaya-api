package dutchman

import (
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/dutchman/models"
	"github.com/sirupsen/logrus"
	"net"
)

type Config struct {
	HttpHost        string `json:"http_host"`
	HttpPort        string `json:"http_port"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

type Dutchman struct {
	cfg Config
	r   *gin.Engine
}

func NewDutchman(cfg Config) (*Dutchman, error) {
	r := gin.Default()

	r.POST("/api/auth/login", func(c *gin.Context) {

		email := c.PostForm("email")
		password := c.PostForm("password")

		user := &models.User{
			ID:       "test",
			Name:     "Egor",
			Email:    email,
			Password: password,
		}
		token, err := GenerateToken(user)
		if err != nil {
			logrus.Errorf("error generating token: %v", err)
			c.AbortWithError(500, err)
			return
		}

		c.JSON(200, gin.H{
			"success": true,
			"token":   token,
		})
	})
	r.GET("/api/auth/user", func(c *gin.Context) {
		user := &models.User{
			ID:       "test",
			Name:     "Egor",
			Email:    "test@test.ru",
			Password: "test@test.ru",
		}

		c.JSON(200, user)
	})

	return &Dutchman{
		cfg: cfg,
		r:   r,
	}, nil
}

func (d *Dutchman) Start() error {
	err := d.r.Run(net.JoinHostPort(d.cfg.HttpHost, d.cfg.HttpPort))
	if err != nil {
		logrus.Fatalf("error running gin: %v", err)
	}

	return nil
}
