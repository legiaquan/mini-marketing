package main

import (
	"fmt"
	"mini-marketing/config"
)

func main() {
	// Khởi tạo cấu hình hệ thống
	config.InitConfig()

	fmt.Println("Hello Mini Marketing")
}
