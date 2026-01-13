package config

type Config struct {
	ServerAddress string
}

func GetConfig() *Config {
	return &Config{ServerAddress: "localhost:8080"}
}
