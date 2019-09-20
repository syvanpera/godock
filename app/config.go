package app

import (
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	// ConfigFolder default name
	ConfigFolder = ".godock"

	// ConfigFile default name
	ConfigFile = "config.toml"

	// ConfigLogFile default name
	ConfigLogFile = "godock.log"
)

type Config struct {
	Godock   *godock
	Flowdock *auth
}

type godock struct {
	Debug   bool
	LogFile string
}

type auth struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	RedirectURL  string
}

func Default() *Config {
	return &Config{
		Godock: &godock{
			Debug:   false,
			LogFile: ConfigLogFile,
		},
		Flowdock: &auth{
			ClientID:     "",
			ClientSecret: "",
			AuthURL:      "https://www.flowdock.com/oauth/authorize",
			TokenURL:     "https://api.flowdock.com/oauth/token",
			RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		},
	}
}

func LoadConfig() (*Config, error) {
	log.Info().Msg("Loading configuration")

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	config := *Default()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
