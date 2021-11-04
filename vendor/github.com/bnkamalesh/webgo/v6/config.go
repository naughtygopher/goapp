package webgo

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"
)

// Config is used for reading app's configuration from json file
type Config struct {
	// Host is the host on which the server is listening
	Host string `json:"host,omitempty"`
	// Port is the port number where the server has to listen for the HTTP requests
	Port string `json:"port,omitempty"`

	// CertFile is the TLS/SSL certificate file path, required for HTTPS
	CertFile string `json:"certFile,omitempty"`
	// KeyFile is the filepath of private key of the certificate
	KeyFile string `json:"keyFile,omitempty"`
	// HTTPSPort is the port number where the server has to listen for the HTTP requests
	HTTPSPort string `json:"httpsPort,omitempty"`

	// ReadTimeout is the maximum duration for which the server would read a request
	ReadTimeout time.Duration `json:"readTimeout,omitempty"`
	// WriteTimeout is the maximum duration for which the server would try to respond
	WriteTimeout time.Duration `json:"writeTimeout,omitempty"`

	// InsecureSkipVerify is the HTTP certificate verification
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// ShutdownTimeout is the duration in which graceful shutdown is completed
	ShutdownTimeout time.Duration
}

// Load config file from the provided filepath and validate
func (cfg *Config) Load(filepath string) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		LOGHANDLER.Fatal(err)
	}

	err = json.Unmarshal(file, cfg)
	if err != nil {
		LOGHANDLER.Fatal(err)
	}

	err = cfg.Validate()
	if err != nil {
		LOGHANDLER.Fatal(ErrInvalidPort)
	}
}

// Validate the config parsed into the Config struct
func (cfg *Config) Validate() error {
	i, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return ErrInvalidPort
	}

	if i <= 0 || i > 65535 {
		return ErrInvalidPort
	}

	return nil
}
