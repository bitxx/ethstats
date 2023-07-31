package config

type Application struct {
	Name      string
	Contract  string
	Version   string
	Secret    string
	ServerUrl string
}

var ApplicationConfig = new(Application)
