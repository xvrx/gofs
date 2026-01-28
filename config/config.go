package config

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"

	// "github.com/go-redis/redis/v8"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Redis RedisConfig            `yaml:"redis"`
	MySQL map[string]MySQLConfig `yaml:"mysql"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type MySQLConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Database string `yaml:"database"`
}

var (
	AppConfig   Config
	RedisClient *redis.Client
	DB          map[string]*sql.DB
)

func LoadConfig() error {
	configPath := "config/config.yaml"
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(configFile, &AppConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

func InitRedis() error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     AppConfig.Redis.Addr,
		Password: AppConfig.Redis.Password,
		DB:       AppConfig.Redis.DB,
	})

	// Ping to check connection
	_, err := RedisClient.Ping(RedisClient.Context()).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	fmt.Println("Connected to Redis!")
	return nil
}

func InitMySQL() error {
	DB = make(map[string]*sql.DB)
	for name, config := range AppConfig.MySQL {
		// Connect without specifying the database
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true",
			config.User,
			config.Password,
			config.Host,
			config.Port)

		db, err := sql.Open("mysql", dsn)
		if err != nil {
			return fmt.Errorf("failed to open database connection for %s: %w", name, err)
		}

		if err = db.Ping(); err != nil {
			return fmt.Errorf("failed to connect to MySQL for %s: %w", name, err)
		}

		// Create the database if it doesn't exist
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", config.Database))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", config.Database, err)
		}
		db.Close()

		// Connect to the specific database
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			config.User,
			config.Password,
			config.Host,
			config.Port,
			config.Database)

		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return fmt.Errorf("failed to open database connection for %s: %w", name, err)
		}

		if err = db.Ping(); err != nil {
			return fmt.Errorf("failed to connect to MySQL for %s: %w", name, err)
		}

		DB[name] = db
		fmt.Printf("Connected to MySQL database '%s'!\n", name)
	}
	return nil
}
