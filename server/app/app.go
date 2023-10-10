package app

import (
	"ethstats/server/app/model"
	"ethstats/server/app/service"
	"ethstats/server/config"
	"github.com/bitxx/logger"
	"github.com/bitxx/logger/logbase"
	"net/http"
)

type App struct {
	logger  *logbase.Helper
	channel *model.Channel
}

func NewApp() *App {
	channel := &model.Channel{
		MsgPing:    make(chan []byte),
		MsgLatency: make(chan []byte),
		Nodes:      make(map[string][]byte),
		LoginIDs:   make(map[string]string),
	}
	return &App{
		channel: channel,
		logger: logger.NewLogger(
			logger.WithType(config.LoggerConfig.Type),
			logger.WithPath(config.LoggerConfig.Path),
			logger.WithLevel(config.LoggerConfig.Level),
			logger.WithStdout(config.LoggerConfig.Stdout),
			logger.WithCap(config.LoggerConfig.Cap),
		),
	}
}

func (a *App) Start() {
	relay := service.NewRelay(a.channel, a.logger)
	api := service.NewApi(a.channel, a.logger)
	http.HandleFunc("/", relay.HandleRequest)
	http.HandleFunc("/api", api.HandleRequest)
	a.logger.Fatal(http.ListenAndServe(config.ApplicationConfig.Host+":"+config.ApplicationConfig.Port, nil))
}
