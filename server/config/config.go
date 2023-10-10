package config

import (
	"fmt"
	loadconfig "github.com/bitxx/load-config"
	"github.com/bitxx/load-config/source"
	"log"
)

type Config struct {
	Application *Application `yaml:"application"`
	Logger      *Logger      `yaml:"logger"`
	Email       *Email       `yaml:"email"`
	callbacks   []func()
}

func (e *Config) init() {
	e.runCallback()
}

func (e *Config) Init() {
	e.init()
	log.Println("!!!  server config init")
}

func (e *Config) runCallback() {
	for i := range e.callbacks {
		e.callbacks[i]()
	}
}

func (e *Config) OnChange() {
	e.init()
	log.Println("!!! server config change and reload")
}

// Setup 载入配置文件
func Setup(s source.Source,
	fs ...func()) {
	_cfg := &Config{
		Application: ApplicationConfig,
		Logger:      LoggerConfig,
		Email:       EmailConfig,
		callbacks:   fs,
	}
	var err error
	loadconfig.DefaultConfig, err = loadconfig.NewConfig(
		loadconfig.WithSource(s),
		loadconfig.WithEntity(_cfg),
	)
	if err != nil {
		log.Println(fmt.Sprintf("New server config object fail: %s, use default param to start", err.Error()))
		return
	}
	_cfg.Init()
}
