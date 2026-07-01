package services

import (
	"context"
	"fmt"
	"mini-marketing/pb"
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

	// Ví dụ Business Logic:
	fmt.Println("🚀 Nhận được yêu cầu tạo Campaign mới!")
	fmt.Printf("Tên: %s\n", req.GetName())
	fmt.Printf("Nội dung: %s\n", req.GetContent())
	fmt.Printf("Số lượng user mục tiêu: %d\n", req.GetTargetUsers())

	// Tạm thời trả về kết quả giả (Mock data)
	// Sau này ở đây sẽ gọi code lưu vào Database (Giai đoạn 3) và đẩy vào Redis Queue (Giai đoạn 5)

	return &pb.CampaignResponse{
		Id:      "camp_1234563wqweqww2",
		Status:  "CREATED",
		Message: "Tạo chiến dịch thành công!",
	}, nil
}
