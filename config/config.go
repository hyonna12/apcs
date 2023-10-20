package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
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
		Level  string `yaml:"level"`
		Output string `yaml:"output"`
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
			} `yaml:"robot"`
			DeadlockCheckPeriod int `yaml:"deadlock-check-period"`
		} `yaml:"resource"`
		Simulation struct {
			Delay int `yaml:"delay"`
		} `yaml:"simulation"`
		TrayBuffer struct {
			Optimum int   `yaml:"optimum"`
			Init    []int `yaml:"init"`
		} `yaml:"trayBuffer"`
	} `yaml:"plc"`

	Sorting struct {
		State string `yaml:"state"`
	} `yaml:"sorting"`
	Kiosk struct {
		Ad string `yaml:"ad"`
	} `yaml:"kiosk"`
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

/* 로그 custom formatter */
type PlainFormatter struct {
	log.TextFormatter
	TimestampFormat        string
	LevelDesc              []string
	DisableLevelTruncation bool
}

func NewPlainFormatter() *PlainFormatter {
	return &PlainFormatter{
		TimestampFormat:        time.TimeOnly,
		LevelDesc:              []string{"PANIC", "ERROR", "WARN", "Info", "DEBUG", "TRACE"},
		DisableLevelTruncation: false,
	}
}

func (f *PlainFormatter) Format(entry *log.Entry) ([]byte, error) {
	timestamp := fmt.Sprintf("%v", entry.Time.Format(f.TimestampFormat))
	level := ""
	message := ""
	if Config.Logging.Output == "terminal" {
		level = color.GreenString(f.LevelDesc[entry.Level])
		message = color.CyanString(entry.Message)
	} else if Config.Logging.Output == "file" {
		level = f.LevelDesc[entry.Level]
		message = entry.Message
	}
	return []byte(fmt.Sprintf("[%s] %s: '%s'\nfunc=%v file=%v:%v\n", timestamp, level, message, entry.Caller.Func.Name(), entry.Caller.File, entry.Caller.Line)), nil
}
