package models

import "time"

// Campaign đại diện cho bảng campaigns trong cơ sở dữ liệu
type Campaign struct {
	ID        string `gorm:"primaryKey;type:varchar(50)"`
	Name      string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Content   string `gorm:"type:text;not null"`
	Status    string `gorm:"type:varchar(20);default:'CREATED'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
