package process

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/hibiken/asynq"

	"mini-marketing/config"
)

// Dữ liệu giả lập nhận từ Dịch vụ Đặt hàng (Service B)
type OrderCompletedEvent struct {
	OrderID string `json:"order_id"`
	UserID  int    `json:"user_id"`
}

func StartKafkaConsumer() {
	// Gắn thẻ nhân viên (Username/Password)
	mechanism := plain.Mechanism{
		Username: config.AppConfig.KafkaUser,
		Password: config.AppConfig.KafkaPassword,
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{config.AppConfig.KafkaBrokers},
		GroupID:  "mini-marketing-group", // RẤT QUAN TRỌNG: Định danh nhóm đọc
		Topic:    "order.completed",
		MaxBytes: 10e6, // Tối đa 10MB
		Dialer:   dialer,
	})
	
	// Khởi tạo Client để ném Task vào Asynq Queue (Ném vào Bếp)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: config.AppConfig.RedisURL})
	defer asynqClient.Close()

	log.Println("🎧 Đang lắng nghe sự kiện từ Kafka trên topic 'order.completed'...")

	for {
		// BƯỚC 1: Dùng FetchMessage (Chỉ LẤY tin nhắn ra xem, KHÔNG XÁC NHẬN ĐÃ ĐỌC)
		m, err := r.FetchMessage(context.Background())
		if err != nil {
			log.Printf("❌ [KAFKA CONSUMER] Lỗi đọc message từ luồng: %v", err)
			break // Thoát nếu Kafka sập
		}
		
		var event OrderCompletedEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			log.Printf("❌ [KAFKA CONSUMER] Lỗi parse JSON: %v", err)
			// Nếu JSON lỗi rác, ta buộc phải Commit để bỏ qua, nếu không sẽ bị kẹt mãi mãi
			r.CommitMessages(context.Background(), m)
			continue
		}

		log.Printf("🎧 [KAFKA CONSUMER] Hóng được sự kiện hoàn thành đơn '%s' của User %d", event.OrderID, event.UserID)

		payload := SendNotificationPayload{
			CampaignID: fmt.Sprintf("auto-event-order-%s", event.OrderID),
			UserIDs:    []int{event.UserID},
		}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask(TypeSendNotification, payloadBytes)
		
		// BƯỚC 2: Cố gắng đẩy vào Redis. Nếu Redis sập, Retry liên tục không bỏ cuộc!
		maxRetries := 5
		success := false
		for i := 1; i <= maxRetries; i++ {
			_, err = asynqClient.Enqueue(task, asynq.Queue("campaign"))
			if err != nil {
				log.Printf("⚠️ [KAFKA -> REDIS] Lỗi đẩy Task vào Queue (Thử lại %d/%d): %v", i, maxRetries, err)
				time.Sleep(2 * time.Second) // Nghỉ ngơi 2s trước khi thử lại
			} else {
				log.Printf("🚀 [KAFKA -> REDIS] Đã ném Task tri ân User %d sang Redis Queue thành công!", event.UserID)
				success = true
				break
			}
		}

		// BƯỚC 3: Chỉ XÁC NHẬN (Commit) với Kafka khi đã nhét vào Redis an toàn!
		if success {
			if err := r.CommitMessages(context.Background(), m); err != nil {
				log.Printf("❌ [KAFKA CONSUMER] Lỗi Commit Message về Kafka: %v", err)
			}
		} else {
			// Nếu 5 lần vẫn sập, ta Crash luôn App để kỹ sư vào sửa Redis. 
			// Yên tâm vì chưa Commit nên message vẫn còn nguyên trên Kafka!
			log.Fatalf("🚨 [CRITICAL] Redis sập toàn tập, dừng Consumer để bảo toàn Message!")
		}
	}
}
