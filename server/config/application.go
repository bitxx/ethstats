package config

type Application struct {
	Name    string
	Host    string
	Port    string
	Version string
	Secret  string
}

var ApplicationConfig = new(Application)
