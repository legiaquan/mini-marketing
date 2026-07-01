package stores

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	
	"mini-marketing/config"
	"mini-marketing/internal/models"
)

// Biến toàn cục để các nơi khác có thể gọi DB
var DB *gorm.DB

func InitMySQL() {
	var err error
	dsn := config.AppConfig.DatabaseURL

	// 1. Mở kết nối
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Lỗi kết nối MySQL: %v", err)
	}

	// 2. Tự động tạo hoặc cập nhật cấu trúc bảng (Auto Migrate)
	err = DB.AutoMigrate(&models.Campaign{})
	if err != nil {
		log.Fatalf("❌ Lỗi AutoMigrate: %v", err)
	}

	log.Println("✅ Đã kết nối MySQL & AutoMigrate thành công!")
}
