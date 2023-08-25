package config

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Config *Configuration
)

type Configuration struct {
	Profile string

	Database struct {
		DriverName string `yaml:"driver-name"`
		Username   string `yaml:"username"`
		Password   string `yaml:"password"`
		Host       string `yaml:"host"`
		Port       int    `yaml:"port"`
		DbName     string `yaml:"db-name"`
	} `yaml:"database"`

	Logging struct {
		Level string `yaml:"level"`
	} `yaml:"logging"`

	Webserver struct {
		Server struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"server"`
	} `yaml:"webserver"`

	Messenger struct {
		Timeout struct {
			// 메시지를 허브에 전파할 때 대기하는 시간 제한(초)
			Spread int `yaml:"spread"`
			// 메시지 응답을 기다리는 시간 제한(초)
			Response int `yaml:"response"`
		} `yaml:"timeout"`
	} `yaml:"messenger"`

	Plc struct {
		Resource struct {
			Robot struct {
				Number int `yaml:"number"`
				Job    struct {
					Timeout       int `yaml:"timeout"`
					PollingPeriod int `yaml:"polling-period"`
				} `yaml:"job"`
			} `yaml:"robot"`
			DeadLockCheckPeriod int `yaml:"deadlock-check-period"`
		} `yaml:"resource"`
		Simulation struct {
			Delay int `yaml:"delay"`
		} `yaml:"simulation"`
	} `yaml:"plc"`
}

func InitConfig() {
	profile := flag.String("profile", "dev", "실행 Profile 지정. 개발환경:dev, 운영환경:prod")
	flag.Parse()
	log.Infof("[Configuration] Run profile: %s", *profile)

	// config.go 파일 절대 경로 얻기
	_, file, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(file)
	log.Infof(basePath)

	configFilePath := fmt.Sprintf("%s/config_%s.yml", basePath, *profile)

	config := &Configuration{}
	configFilename, err := filepath.Abs(configFilePath)
	if err != nil {
		log.Panic(err)
	}

	configFile, err := os.ReadFile(configFilename)
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Panic(err)
	}

	Config = config
	Config.Profile = *profile
}
