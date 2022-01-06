package global

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/jinzhu/gorm"
	"gopkg.in/yaml.v2"
)

type Config struct {
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
	Gzip            bool   `yaml:"Gzip"`
}

type Environment struct {
	cfg  Config
	db   *gorm.DB
	Gzip bool
}

var env Environment

func initializeEnvironment() {
	var config_file string = "config.yaml"
	flag.StringVar(&config_file, "c", "", "config file")
	config_file = "/root/work/strawberry/bin/config/config.yaml"
	log.Println(config_file)
	result, err := ioutil.ReadFile(config_file)
	if err != nil {
		log.Fatalf("Failed to read config, err: %s", err)
		return
	}
	err = yaml.Unmarshal(result, &env.cfg)
	if err != nil {
		log.Fatalf("Failed to load config, err: %s", err)
		return
	}

	if env.cfg.WebServerAddr == "" {
		port := os.Getenv("GinAddr")
		if port == "" {
			env.cfg.WebServerAddr = ":8080"
			log.Printf("Defaulting to port %s", port)
		}
	}

}

var once sync.Once

func GetEnvironmentConfig() *Environment {
	once.Do(initializeEnvironment)
	return &env
}

func (env *Environment) WebServerAddr() string {
	return env.cfg.WebServerAddr
}

func (env *Environment) WebServerEnv() string {
	return env.cfg.WebServerEnv
}

func (env *Environment) WebRelativePath() string {
	return env.cfg.WebRelativePath
}

func (env *Environment) WebAbsolutePath() string {
	return env.cfg.WebAbsolutePath
}
