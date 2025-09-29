package config

import (
    "log"
    "sync"

    "github.com/spf13/viper"
)

type Config struct {
    Environment string `mapstructure:"ENV"`
    Port        string `mapstructure:"PORT"`
    DatabaseURL string `mapstructure:"DATABASE_URL"`
    LogLevel    string `mapstructure:"LOG_LEVEL"`
}

var (
    cfg  Config
    once sync.Once
)

func Load() Config {
    once.Do(func() {
        viper.SetConfigFile(".env")
        viper.SetConfigType("env")
        viper.AutomaticEnv()

        if err := viper.ReadInConfig(); err != nil {
            log.Printf("config: proceeding without .env file: %v", err)
        }

        // Defaults
        viper.SetDefault("ENV", "development")
        viper.SetDefault("PORT", "8080")
        viper.SetDefault("LOG_LEVEL", "info")

        if err := viper.Unmarshal(&cfg); err != nil {
            log.Fatalf("config: unable to decode into struct: %v", err)
        }
    })

    return cfg
}

