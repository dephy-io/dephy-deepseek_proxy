package config

import (
    "fmt"
    "log"
    "os"

    "gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
    Database struct {
        Host     string `yaml:"host"`
        User     string `yaml:"user"`
        Password string `yaml:"password"`
        DBName   string `yaml:"dbname"`
        Port     string `yaml:"port"`
        SSLMode  string `yaml:"sslmode"`
    } `yaml:"database"`
    Nostr struct {
        RelayURL    string `yaml:"relay_url"`
        MachinePubkey string `yaml:"machine_pubkey"`
    } `yaml:"nostr"`
    Chat struct {
        APIKey string `yaml:"api_key"`
		MaxContextTokens int `yaml:"max_context_tokens"`
    } `yaml:"chat"`
    Server struct {
        Port int `yaml:"port"` 
    } `yaml:"server"`
}

// GlobalConfig is the global configuration instance
var GlobalConfig Config

// DSN generates the PostgreSQL DSN from database config
func (c *Config) DSN() string {
    return fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
        c.Database.Host,
        c.Database.User,
        c.Database.Password,
        c.Database.DBName,
        c.Database.Port,
        c.Database.SSLMode,
    )
}

// LoadConfig reads and parses the YAML configuration file into GlobalConfig
func LoadConfig(filePath string) error {
    // Read the YAML file
    data, err := os.ReadFile(filePath)
    if err != nil {
        return err
    }

    // Unmarshal YAML into GlobalConfig
    if err := yaml.Unmarshal(data, &GlobalConfig); err != nil {
        return err
    }

    // Validate required fields
    if GlobalConfig.Database.Host == "" {
        log.Fatal("database.host is required in config.yaml")
    }
    if GlobalConfig.Database.User == "" {
        log.Fatal("database.user is required in config.yaml")
    }
    if GlobalConfig.Database.Password == "" {
        log.Fatal("database.password is required in config.yaml")
    }
    if GlobalConfig.Database.DBName == "" {
        log.Fatal("database.dbname is required in config.yaml")
    }
    if GlobalConfig.Database.Port == "" {
        log.Fatal("database.port is required in config.yaml")
    }
    if GlobalConfig.Database.SSLMode == "" {
        log.Fatal("database.sslmode is required in config.yaml")
    }
    if GlobalConfig.Nostr.RelayURL == "" {
        log.Fatal("nostr.relay_url is required in config.yaml")
    }
    if GlobalConfig.Nostr.MachinePubkey == "" {
        log.Fatal("nostr.machine_pubkey is required in config.yaml")
    }
    if GlobalConfig.Chat.APIKey == "" {
        log.Fatal("chat.api_key is required in config.yaml")
    }
    if GlobalConfig.Server.Port == 0 {
        log.Fatal("server.port is required in config.yaml")
    }
	if GlobalConfig.Server.Port < 1 || GlobalConfig.Server.Port > 65535 {
		log.Fatal("server.port must be between 1 and 65535")
	}

    return nil
}