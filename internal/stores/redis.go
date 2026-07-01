package stores

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	
	"mini-marketing/config"
)

// Biến toàn cục để tương tác với Redis
var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: config.AppConfig.RedisURL,
	})

	// PING thử xem có kết nối được không
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("❌ Lỗi kết nối Redis: %v", err)
	}

	log.Println("✅ Đã kết nối Redis thành công!")
}
