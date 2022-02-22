package dutchman

type Config struct {
	HttpHost        string `json:"http_host"`
	HttpPort        string `json:"http_port"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

type Dutchman struct {
	cfg Config
}

func NewDutchman(cfg Config) (*Dutchman, error) {
	return &Dutchman{}, nil
}

func (d *Dutchman) Start() error {
	return nil
}
