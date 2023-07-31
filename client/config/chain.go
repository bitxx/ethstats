package config

type Chain struct {
	Url     string
	Timeout int64
	Port    string
}

var ChainConfig = new(Chain)
