package config

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type GlobalConfig struct {
	WebServerEnv    string `yaml:"web-server-env"`
	WebServerAddr   string `yaml:"web-server-addr"`
	DataBaseType    string `yaml:"database-type"`
	DataBaseAddr    string `yaml:"database-addr"`
	DataBasePort    int    `yaml:"database-port"`
	WebRelativePath string `yaml:"web-relative-path"`
	WebAbsolutePath string `yaml:"web-absolute-path"`
	WebUploadPath   string `yaml:"web-upload-path"`
	WebDownloadPath string `yaml:"web-download-path"`
	LoggerPath      string `yaml:"logger-path"`
	LoggerLevel     string `yaml:"logger-level"`
}

var cfg *GlobalConfig = nil

func InitializeGlobalConfig() *GlobalConfig {
	result, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Failed to read configuration")
		return nil
	}
	err = yaml.Unmarshal(result, &cfg)
	if err != nil {
		log.Fatalf("Failed to load configuration")
		return nil
	}

	if cfg.WebServerAddr == "" {
		port := os.Getenv("GinAddr")
		if port == "" {
			cfg.WebServerAddr = ":8080"
			log.Printf("Defaulting to port %s", port)
		}
	}

	return cfg
}

func GetGlobalConfig() *GlobalConfig {
	if cfg == nil {
		cfg = InitializeGlobalConfig()
	}
	return cfg
}
