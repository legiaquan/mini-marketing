package services

import (
	"context"
	"encoding/json"
	"fmt"
	
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	
	"mini-marketing/config"
	"mini-marketing/pb"
	"mini-marketing/internal/models"
	"mini-marketing/internal/stores"
	"mini-marketing/internal/process"
)

// 1. Khai báo một Struct đại diện cho Service của bạn
// Nó phải "kế thừa" (embed) UnimplementedMarketingServiceServer bắt buộc từ file gRPC sinh ra.
type CampaignService struct {
	pb.UnimplementedMarketingServiceServer
}

// 2. Viết hàm tạo mới Service (Constructor)
func NewCampaignService() *CampaignService {
	return &CampaignService{}
}

// 3. ĐÂY CHÍNH LÀ NƠI VIẾT BUSINESS LOGIC!
// Hàm này phải có tên và tham số khớp 100% với định nghĩa trong file `service_grpc.pb.go`.
func (s *CampaignService) CreateCampaign(ctx context.Context, req *pb.CampaignRequest) (*pb.CampaignResponse, error) {

	fmt.Println("🚀 Nhận được yêu cầu tạo Campaign mới!")

	// 0.1 Validation: Sử dụng hàm sinh tự động từ file .proto (protoc-gen-validate)
	if err := req.ValidateAll(); err != nil {
		fmt.Printf("⚠️ Lỗi Validation: %v\n", err)
		// Trả về lỗi 400 kèm theo chi tiết trường nào bị sai
		return nil, status.Errorf(codes.InvalidArgument, "Dữ liệu không hợp lệ: %v", err)
	}

	// 0.2 Validation: Kiểm tra trùng tên Campaign (Business Validation)
	var count int64
	stores.DB.Model(&models.Campaign{}).Where("name = ?", req.GetName()).Count(&count)
	if count > 0 {
		fmt.Printf("⚠️ Lỗi: Tên '%s' đã tồn tại!\n", req.GetName())
		// Trả về mã lỗi chuẩn của gRPC (409 Conflict ở HTTP)
		return nil, status.Errorf(codes.AlreadyExists, "Tên chiến dịch '%s' đã tồn tại, vui lòng chọn tên khác", req.GetName())
	}

	// 1. Tạo Model dữ liệu để chuẩn bị lưu vào Database
	newCampaign := models.Campaign{
		ID:      uuid.New().String(),
		Name:    req.GetName(),
		Content: req.GetContent(),
		Status:  "CREATED", // Trạng thái mặc định
	}

	// 2. Gọi tầng stores (GORM) để INSERT vào MySQL
	if err := stores.DB.Create(&newCampaign).Error; err != nil {
		fmt.Printf("❌ Lỗi lưu Database: %v\n", err)
		return nil, fmt.Errorf("không thể tạo campaign lúc này")
	}

	fmt.Printf("✅ Đã lưu Campaign '%s' vào Database thành công!\n", newCampaign.ID)

	// 2.5 Giai đoạn 5: Đẩy Task vào Queue để chạy ngầm
	// Tạo danh sách User IDs giả lập theo số lượng khách hàng yêu cầu
	var mockUserIDs []int
	for i := 1; i <= int(req.GetTargetUsers()); i++ {
		mockUserIDs = append(mockUserIDs, i)
	}

	payloadBytes, _ := json.Marshal(process.SendNotificationPayload{
		CampaignID: newCampaign.ID,
		UserIDs:    mockUserIDs,
	})

	task := asynq.NewTask(process.TypeSendNotification, payloadBytes)
	
	// Khởi tạo Asynq Client (Thực tế nên gom thành biến toàn cục ở stores, nhưng để gọn ta khởi tạo tạm ở đây)
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: config.AppConfig.RedisURL})
	defer asynqClient.Close()

	// Chỉ định đẩy Task vào Queue có tên là "campaign" (thay vì "default")
	info, err := asynqClient.Enqueue(task, asynq.Queue("campaign"))
	if err != nil {
		fmt.Printf("❌ Lỗi đẩy Task vào Queue: %v\n", err)
	} else {
		fmt.Printf("📦 Đã đẩy Task vào Queue! TaskID: %s\n", info.ID)
	}

	// 3. Trả về kết quả cho người dùng (Không bị đứng chờ gửi thông báo xong)
	return &pb.CampaignResponse{
		Id:      newCampaign.ID,
		Status:  newCampaign.Status,
		Message: "Tạo chiến dịch và lưu Database thành công!",
	}, nil
}
