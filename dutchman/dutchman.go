package dutchman

import (
	"github.com/gin-gonic/gin"
	"github.com/lightswitch/dutchman-backend/dutchman/database"
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

	r  *gin.Engine
	db *database.Database
}

func NewDutchman(cfg Config) (*Dutchman, error) {
	d := &Dutchman{
		cfg: cfg,
	}

	// Gin server
	r := gin.Default()
	d.registerRoutes(r)
	d.r = r

	// Database
	db, err := database.NewDatabase(database.Config{})
	if err != nil {
		logrus.Fatalf("error creating database: %v", err)
	}
	d.db = db

	//logrus.Info(db.GetUser("c0c6c001-9fdd-499c-84d3-051cbdcd9cfb"))

	return d, nil
}

func (d *Dutchman) Start() error {
	err := d.r.Run(net.JoinHostPort(d.cfg.HttpHost, d.cfg.HttpPort))
	if err != nil {
		logrus.Fatalf("error running gin: %v", err)
	}

	return nil
}
