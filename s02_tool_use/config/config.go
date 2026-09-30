package config

import "os"

var LocalConfig Config

type Config struct {
	APIKey     string
	BaseUrl    string
	Model      string
	LccWorkDir string
}

func LoadConfig() Config {
	return Config{
		APIKey:     os.Getenv("DEEPSEEK_API_KEY"),
		BaseUrl:    os.Getenv("DEEPSEEK_API_BASE_URL"),
		Model:      os.Getenv("DEEPSEEK_API_MODEL"),
		LccWorkDir: os.Getenv("LCC_WORKDIR"),
	}
}

func init() {
	LocalConfig = LoadConfig()
}
