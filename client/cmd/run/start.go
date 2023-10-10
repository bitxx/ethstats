package run

import (
	"ethstats/client/app"
	"ethstats/client/config"
	"github.com/bitxx/load-config/source/file"
	"github.com/spf13/cobra"
	"log"
)

var (
	configPath string
	StartCmd   *cobra.Command
)

const (
	name      = "name"
	contract  = "contract"
	version   = "version"
	secret    = "secret"
	serverUrl = "server-url"
	logPath   = "log-path"
	logLevel  = "log-level"
	logStdout = "log-stdout"
	logType   = "log-type"
	logCap    = "log-cap"
	chainUrl  = "chain-url"
	chainPort = "chain-port"
)

func init() {
	StartCmd = &cobra.Command{
		Use:          "start",
		Short:        "run the client",
		Example:      "client start -c config/settings.yml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			config.Setup(
				file.NewSource(file.WithPath(configPath)),
			)
			flag := cmd.PersistentFlags()

			if name, _ := flag.GetString(name); name != "" {
				config.ApplicationConfig.Name = name
			}
			if contract, _ := flag.GetString(contract); contract != "" {
				config.ApplicationConfig.Contract = contract
			}
			if version, _ := flag.GetString(version); version != "" && config.ApplicationConfig.Version == "" {
				config.ApplicationConfig.Version = version
			}
			if secret, _ := flag.GetString(secret); secret != "" && config.ApplicationConfig.Secret == "" {
				config.ApplicationConfig.Secret = secret
			}
			if serverUrl, _ := flag.GetString(serverUrl); serverUrl != "" && config.ApplicationConfig.ServerUrl == "" {
				config.ApplicationConfig.ServerUrl = serverUrl
			}
			if logPath, _ := flag.GetString(logPath); logPath != "" && config.LoggerConfig.Path == "" {
				config.LoggerConfig.Path = logPath
			}
			if logLevel, _ := flag.GetString(logLevel); logLevel != "" && config.LoggerConfig.Level == "" {
				config.LoggerConfig.Level = logLevel
			}
			if logStdout, _ := flag.GetString(logStdout); logStdout != "" && config.LoggerConfig.Stdout == "" {
				config.LoggerConfig.Stdout = logStdout
			}
			if logType, _ := flag.GetString(logType); logType != "" && config.LoggerConfig.Type == "" {
				config.LoggerConfig.Type = logType
			}
			if logCap, _ := flag.GetUint(logCap); logCap > 0 && config.LoggerConfig.Cap <= 0 {
				config.LoggerConfig.Cap = logCap
			}
			if chainUrl, _ := flag.GetString(chainUrl); chainUrl != "" {
				config.ChainConfig.Url = chainUrl
			}
			if chainPort, _ := flag.GetString(chainPort); chainPort != "" && config.ChainConfig.Port == "" {
				config.ChainConfig.Port = chainPort
			}
			if config.ApplicationConfig.Name == "" {
				log.Fatal("param name can't empty")
			}
			if config.ApplicationConfig.Secret == "" {
				log.Fatal("param secret can't empty")
			}
			if config.ApplicationConfig.ServerUrl == "" {
				log.Fatal("param serverUrl can't empty")
			}
			if config.ChainConfig.Url == "" {
				log.Fatal("param chainUrl can't empty")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			return run()
		},
	}
	cmd := StartCmd.PersistentFlags()
	cmd.StringVarP(&configPath, "config", "c", "", "Start server with provided configuration file")
	cmd.String(name, "", "node name")
	cmd.String(contract, "", "contract")
	cmd.String(version, "v1.0.0", "version")
	cmd.String(secret, "", "secret")
	cmd.String(serverUrl, "", "server url")
	cmd.String(logPath, "", "log path")
	cmd.String(logLevel, "trace", "log level")
	cmd.String(logStdout, "default", "default,file")
	cmd.String(logType, "default", "default、zap、logrus")
	cmd.Uint(logCap, 50, "log cap")
	cmd.String(chainUrl, "", "chain url with port,eg:https://127.0.0.1:30303")
	cmd.String(chainPort, "30303", "chain port,use for report")
}

func run() error {
	app.NewApp().Start()
	return nil
}
