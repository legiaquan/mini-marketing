package process

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"mini-marketing/config"
	"mini-marketing/internal/models"
	"mini-marketing/internal/stores"
)

// 1. Hàm xử lý logic chính khi bốc được Task ra khỏi Queue
func HandleSendNotificationTask(ctx context.Context, t *asynq.Task) error {
	var payload SendNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("📥 [REDIS WORKER] Bắt đầu lấy Task ra khỏi Queue (Campaign: '%s', Tổng Users: %d)", payload.CampaignID, len(payload.UserIDs))

	// [MỚI] Báo cho MySQL biết: Đầu bếp đã bắt đầu xào phở
	// LƯU Ý: Ở dự án thật, ta phải INSERT dòng Campaign này vào MySQL ở bước Kafka Consumer trước.
	// Hiện tại GORM chạy Update("status") nhưng không có báo lỗi (chỉ trả về 0 RowsAffected) do MySQL cho phép Update trên bản ghi không tồn tại.
	stores.DB.Model(&models.Campaign{}).Where("id = ?", payload.CampaignID).Update("status", "PROCESSING")

	// GIẢ LẬP LỖI: Random 30% xác suất tác vụ này sẽ bị văng lỗi (Rớt mạng/Sập DB)
	// if rand.Intn(100) < 30 {
	// 	log.Printf("🔥 [Worker] CẢNH BÁO: Rớt mạng đột ngột khi xử lý Campaign '%s'! (Asynq sẽ tự động Retry)", payload.CampaignID)
	// 	return fmt.Errorf("kết nối đến máy chủ Firebase bị gián đoạn")
	// }

	// Vòng lặp giả lập việc gửi thông báo
	for _, userID := range payload.UserIDs {
		// Giả lập độ trễ mạng khi gọi sang Firebase/APNs
		time.Sleep(10 * time.Millisecond)
		log.Printf("📨 [REDIS WORKER] Đã gọi API Firebase gửi Push Noti thành công đến UserID: %d", userID)

		// [MỚI] Tăng biến đếm trong Redis lên 1 đơn vị cực kỳ nhanh chóng
		redisKey := fmt.Sprintf("campaign:%s:processed", payload.CampaignID)
		stores.RedisClient.Incr(ctx, redisKey)
	}

	// [MỚI] Báo cho MySQL biết: Đã nấu xong 100 tô phở thành công!
	stores.DB.Model(&models.Campaign{}).Where("id = ?", payload.CampaignID).Update("status", "COMPLETED")

	log.Printf("✅ [REDIS WORKER] Hoàn thành toàn bộ quy trình cho Campaign '%s'", payload.CampaignID)
	return nil
}

// 2. Hàm khởi động con Worker đứng canh Queue ngày đêm
func RunWorkerServer() {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.AppConfig.RedisURL},
		asynq.Config{
			Concurrency: 5, // Số lượng luồng xử lý đồng thời cực đại
			Queues: map[string]int{
				"campaign": 1, // Tên queue là 'campaign', độ ưu tiên là 1
			},
		},
	)

	// Đăng ký Handler (khi gặp loại Task này thì gọi hàm nào)
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeSendNotification, HandleSendNotificationTask)

	log.Println("🚀 Bắt đầu khởi động Asynq Worker Server...")
	if err := srv.Run(mux); err != nil {
		log.Fatalf("❌ Lỗi khởi động Worker: %v", err)
	}
}
