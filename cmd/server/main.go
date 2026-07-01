package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"google.golang.org/grpc/reflection"

	"mini-marketing/config"
	"mini-marketing/internal/services"
	"mini-marketing/internal/stores"
	"mini-marketing/pb"
)

func main() {
	// 1. Khởi tạo cấu hình hệ thống
	config.InitConfig()
	fmt.Println("🚀 Bắt đầu khởi động Mini Marketing Server...")

	// Khởi tạo kết nối Databases
	stores.InitMySQL()
	stores.InitRedis()

	// 2. Khởi tạo Service logic của chúng ta
	campaignService := services.NewCampaignService()

	// 3. Khởi tạo gRPC Server (chạy ở cổng 9090)
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Lỗi không thể mở cổng 9090: %v", err)
	}
	grpcServer := grpc.NewServer()
	
	// Đăng ký Service vào gRPC Server
	pb.RegisterMarketingServiceServer(grpcServer, campaignService)

	// Bật Reflection: Giúp các tool test (Postman, grpcurl) tự động quét được API mà không cần file .proto
	reflection.Register(grpcServer)

	// Chạy gRPC server ngầm (Goroutine) để không block luồng chính
	go func() {
		fmt.Println("✅ gRPC Server đang chạy ở cổng :9090")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Lỗi chạy gRPC server: %v", err)
		}
	}()

	// 4. Khởi tạo HTTP Gateway Server (chạy ở cổng 8080) để Postman gọi được
	// Tạo bộ định tuyến (Mux)
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	
	// Đăng ký Gateway, yêu cầu nó trỏ ngược lại cổng gRPC 9090
	err = pb.RegisterMarketingServiceHandlerFromEndpoint(context.Background(), mux, "localhost:9090", opts)
	if err != nil {
		log.Fatalf("Lỗi đăng ký HTTP Gateway: %v", err)
	}

	fmt.Printf("✅ HTTP REST Server đang chạy ở cổng :%s\n", config.AppConfig.Port)
	// Chạy HTTP Server (Lệnh này sẽ block và giữ chương trình tiếp tục chạy)
	if err := http.ListenAndServe(":"+config.AppConfig.Port, mux); err != nil {
		log.Fatalf("Lỗi chạy HTTP server: %v", err)
	}
}
