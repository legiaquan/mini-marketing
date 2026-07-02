package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

func main() {
	// Gắn thẻ nhân viên (Username/Password)
	mechanism := plain.Mechanism{
		Username: "admin",
		Password: "secret_password",
	}

	sharedTransport := &kafka.Transport{
		SASL: mechanism,
	}

	// Khởi tạo một Nhà xuất bản (Producer) đóng vai trò là Dịch vụ Đặt hàng (Service B)
	w := &kafka.Writer{
		Addr:                   kafka.TCP("localhost:9092"),
		Topic:                  "order.completed",
		Balancer:               &kafka.LeastBytes{},
		Transport:              sharedTransport,
		AllowAutoTopicCreation: true, // Cho phép tự tạo kênh nếu kênh chưa tồn tại
	}
	defer w.Close()

	// Khởi tạo random seed
	rand.Seed(time.Now().UnixNano())

	// Sử dụng WaitGroup để chờ tất cả các luồng hoàn thành (Giống Promise.all trong NodeJS)
	var wg sync.WaitGroup

	fmt.Println("🚀 Bắt đầu xả đạn! Đang bắn 10 sự kiện cùng lúc lên Kafka...")

	for i := 0; i < 10; i++ {
		wg.Add(1)

		// Mỗi vòng lặp sẽ chạy một Goroutine độc lập (Không chờ đợi nhau)
		go func(index int) {
			defer wg.Done()

			// Random UserID từ 1 đến 1000, OrderID thêm số ngẫu nhiên
			randomUserID := rand.Intn(1000) + 1
			randomOrderID := fmt.Sprintf("ORD-%d", rand.Intn(999999)+100000)

			message := fmt.Sprintf(`{"order_id": "%s", "user_id": %d}`, randomOrderID, randomUserID)

			// Bắn tin nhắn lên Kafka
			err := w.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(fmt.Sprintf("Key-User-%d", randomUserID)),
					Value: []byte(message),
				},
			)

			if err != nil {
				log.Printf("❌ Lỗi gửi tin nhắn thứ %d: %v", index+1, err)
			} else {
				log.Printf("✅ Đã bắn thành công tin nhắn thứ %d (User: %d)", index+1, randomUserID)
			}
		}(i)
	}

	// Đứng đợi tất cả các Goroutine chạy xong mới kết thúc chương trình
	wg.Wait()
	fmt.Println("🎉 BÙM! Đã xả đạn xong 10 sự kiện!")
}
