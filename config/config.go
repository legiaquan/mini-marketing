package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Port         string `mapstructure:"PORT"`
	DatabaseURL  string `mapstructure:"DB_URL"`
	RedisURL     string `mapstructure:"REDIS_URL"`
	KafkaBrokers  string `mapstructure:"KAFKA_BROKERS"`
	KafkaUser     string `mapstructure:"KAFKA_USER"`
	KafkaPassword string `mapstructure:"KAFKA_PASSWORD"`
	IsProcess     bool   `mapstructure:"IS_PROCESS"`
}

var AppConfig Config

func InitConfig() {
	viper.AutomaticEnv()

	// Cho phép load các biến môi trường có dấu chấm hoặc gạch ngang thay bằng dấu gạch dưới
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Cấu hình các giá trị mặc định (Default values) nếu không truyền biến môi trường
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("DB_URL", "user:password@tcp(localhost:3306)/mini_marketing?parseTime=true")
	viper.SetDefault("REDIS_URL", "localhost:6379")
	viper.SetDefault("KAFKA_BROKERS", "localhost:9092")
	viper.SetDefault("KAFKA_USER", "admin")
	viper.SetDefault("KAFKA_PASSWORD", "secret_password")
	viper.SetDefault("IS_PROCESS", false)

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}

	fmt.Println("Config loaded successfully!")
	fmt.Printf("App mode (IS_PROCESS): %v\n", AppConfig.IsProcess)
}
