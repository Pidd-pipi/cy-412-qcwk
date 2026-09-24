package model

import "time"

type Announcement struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	PublisherID uint      `json:"publisher_id"`
	Publisher   User      `json:"publisher"`
	PublishAt   time.Time `json:"publish_at"`
	Top         bool      `json:"top"`
	ReadCount   int       `json:"read_count"`
	// 发布范围：all=全小区，building=指定楼栋，unit=指定单元
	Scope    string `gorm:"size:20;default:all;index" json:"scope"`
	Building string `gorm:"size:30;index" json:"building"`
	Unit     string `gorm:"size:30;index" json:"unit"`

	// 非持久化字段，由 service 按当前查看者填充。
	// 当前住户是否已确认（紧急公告），住户列表/详情使用。
	Confirmed bool `gorm:"-" json:"confirmed,omitempty"`
	// 物业查看紧急公告时的统计与未确认名单。
	ConfirmedCount   int    `gorm:"-" json:"confirmed_count,omitempty"`
	UnconfirmedUsers []User `gorm:"-" json:"unconfirmed_users,omitempty"`
}
type AnnouncementRead struct {
	ID             uint `gorm:"primaryKey"`
	AnnouncementID uint `gorm:"uniqueIndex:idx_read"`
	UserID         uint `gorm:"uniqueIndex:idx_read"`
	// Confirmed 为 true 表示住户对紧急公告点击了“确认已读”。
	Confirmed   bool       `gorm:"default:false" json:"confirmed"`
	CreatedAt   time.Time  `json:"created_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty"`
}
