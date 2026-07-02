package process

// Định nghĩa tên của Task (giống như chủ đề của tờ giấy ghi chú)
const (
	TypeSendNotification = "notification:send"
)

// Định nghĩa dữ liệu sẽ được gói vào Task (Payload)
type SendNotificationPayload struct {
	CampaignID string
	UserIDs    []int
}
