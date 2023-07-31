package run

import (
	"ethstats/server/app"
	"ethstats/server/config"
	"github.com/bitxx/load-config/source/file"
	"github.com/spf13/cobra"
	"log"
)

var (
	configPath string
	StartCmd   *cobra.Command
)

const (
	name               = "name"
	host               = "host"
	port               = "port"
	version            = "version"
	secret             = "secret"
	logPath            = "log-path"
	logLevel           = "log-level"
	logStdout          = "log-stdout"
	logType            = "log-type"
	logCap             = "log-cap"
	emailHost          = "email-host"
	emailPort          = "email-port"
	emailUsername      = "email-username"
	emailPassword      = "email-password"
	emailFrom          = "email-from"
	emailContentType   = "email-content-type"
	emailTo            = "email-to"
	emailSubjectPrefix = "email-subject-prefix"
	monitorTime        = "email-monitor-time"
)

func init() {
	StartCmd = &cobra.Command{
		Use:          "start",
		Short:        "run the server",
		Example:      "server start -c config/settings.yml",
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			config.Setup(
				file.NewSource(file.WithPath(configPath)),
			)
			flag := cmd.PersistentFlags()

			if name, _ := flag.GetString(name); name != "" {
				config.ApplicationConfig.Name = name
			}
			if host, _ := flag.GetString(host); host != "" {
				config.ApplicationConfig.Host = host
			}
			if port, _ := flag.GetString(port); port != "" {
				config.ApplicationConfig.Port = port
			}
			if version, _ := flag.GetString(version); version != "" && config.ApplicationConfig.Version == "" {
				config.ApplicationConfig.Version = version
			}
			if secret, _ := flag.GetString(secret); secret != "" && config.ApplicationConfig.Secret == "" {
				config.ApplicationConfig.Secret = secret
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
			if emailHost, _ := flag.GetString(emailHost); emailHost != "" && config.EmailConfig.Host == "" {
				config.EmailConfig.Host = emailHost
			}
			if emailPort, _ := flag.GetInt(emailPort); emailPort > 0 && config.EmailConfig.Port <= 0 {
				config.EmailConfig.Port = emailPort
			}
			if emailContentType, _ := flag.GetString(emailContentType); emailContentType != "" && config.EmailConfig.ContentType == "" {
				config.EmailConfig.ContentType = emailContentType
			}
			if emailUsername, _ := flag.GetString(emailUsername); emailUsername != "" && config.EmailConfig.Username == "" {
				config.EmailConfig.Username = emailUsername
			}
			if emailPassword, _ := flag.GetString(emailPassword); emailPassword != "" && config.EmailConfig.Password == "" {
				config.EmailConfig.Password = emailPassword
			}
			if emailFrom, _ := flag.GetString(emailFrom); emailFrom != "" && config.EmailConfig.FromEmail == "" {
				config.EmailConfig.FromEmail = emailFrom
			}
			if emailTo, _ := flag.GetString(emailTo); emailTo != "" && config.EmailConfig.ToEmail == "" {
				config.EmailConfig.ToEmail = emailTo
			}
			if emailSubjectPrefix, _ := flag.GetString(emailSubjectPrefix); emailSubjectPrefix != "" && config.EmailConfig.SubjectPrefix == "" {
				config.EmailConfig.SubjectPrefix = emailSubjectPrefix
			}
			if monitorTime, _ := flag.GetInt(monitorTime); monitorTime > 0 && config.EmailConfig.MonitorTime <= 0 {
				config.EmailConfig.MonitorTime = monitorTime
			}

			if config.ApplicationConfig.Name == "" {
				log.Fatal("param name can't empty")
			}
			if config.ApplicationConfig.Host == "" {
				log.Fatal("param host can't empty")
			}
			if config.ApplicationConfig.Port == "" {
				log.Fatal("param port can't empty")
			}
			if config.ApplicationConfig.Secret == "" {
				log.Fatal("param secret can't empty")
			}

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}
	cmd := StartCmd.PersistentFlags()
	cmd.StringVarP(&configPath, "config", "c", "", "Start server with provided configuration file")

	cmd.String(name, "", "name")
	cmd.String(host, "", "host")
	cmd.String(port, "", "prot")
	cmd.String(version, "v1.0.0", "version")
	cmd.String(secret, "", "secret")
	cmd.String(logPath, "files/logs", "log path")
	cmd.String(logLevel, "trace", "log path")
	cmd.String(logStdout, "file", "default,file")
	cmd.String(logType, "default", "default、zap、logrus")
	cmd.Uint(logCap, 50, "log cap")
	cmd.String(emailHost, "", "email host")
	cmd.Int(emailPort, 0, "email port")
	cmd.String(emailContentType, "text/plain", "email content type")
	cmd.String(emailUsername, "", "email username")
	cmd.String(emailPassword, "", "email password")
	cmd.String(emailFrom, "", "email from")
	cmd.String(emailTo, "", "email to")
	cmd.String(emailSubjectPrefix, "", "email subject prefix")
	cmd.Int(monitorTime, 86400, "email monitor time")
}

func run() error {
	app.NewApp().Start()
	return nil
}
