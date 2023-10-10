package config

type Logger struct {
	Type      string
	Path      string
	Level     string
	Stdout    string
	EnabledDB bool
	Cap       uint
}

var LoggerConfig = new(Logger)
