package model

import "time"

const (
	AnnouncementScopeAll      = "all"
	AnnouncementScopeBuilding = "building"
	AnnouncementScopeUnit     = "unit"
)

type Announcement struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	Category       string    `json:"category"`
	Scope          string    `gorm:"size:20;default:all" json:"scope"`
	ScopeBuilding  string    `gorm:"size:50" json:"scope_building"`
	ScopeUnit      string    `gorm:"size:50" json:"scope_unit"`
	PublisherID    uint      `json:"publisher_id"`
	Publisher      User      `json:"publisher"`
	PublishAt      time.Time `json:"publish_at"`
	Top            bool      `json:"top"`
	ReadCount      int       `json:"read_count"`
	ConfirmedCount int       `json:"confirmed_count"`
	// Confirmed records whether the current resident already confirmed
	// an urgent announcement; staff responses leave it unset.
	Confirmed bool `gorm:"-" json:"confirmed"`
	// UnconfirmedUsers lists residents inside the target scope who have
	// not yet confirmed; only populated for staff viewing 紧急 announcements.
	UnconfirmedUsers []ShortUser `gorm:"-" json:"unconfirmed_users"`
}

// ShortUser is the compact resident profile used by confirmation lists.
type ShortUser struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Building string `json:"building"`
	Unit     string `json:"unit"`
	Room     string `json:"room"`
}

type AnnouncementRead struct {
	ID             uint `gorm:"primaryKey"`
	AnnouncementID uint `gorm:"uniqueIndex:idx_read"`
	UserID         uint `gorm:"uniqueIndex:idx_read"`
	CreatedAt      time.Time
}

// AnnouncementConfirm stores explicit 已读确认 of 紧急 announcements.
// The unique index makes repeated confirmations idempotent.
type AnnouncementConfirm struct {
	ID             uint `gorm:"primaryKey"`
	AnnouncementID uint `gorm:"uniqueIndex:idx_confirm"`
	UserID         uint `gorm:"uniqueIndex:idx_confirm"`
	CreatedAt      time.Time
}
