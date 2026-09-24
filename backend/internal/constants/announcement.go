package constants

// 公告发布范围：全小区 / 指定楼栋 / 指定单元
const (
	AnnouncementScopeAll      = "all"
	AnnouncementScopeBuilding = "building"
	AnnouncementScopeUnit     = "unit"
)

// ValidAnnouncementScopes 发布公告时允许选择的范围。
var ValidAnnouncementScopes = map[string]bool{
	AnnouncementScopeAll:      true,
	AnnouncementScopeBuilding: true,
	AnnouncementScopeUnit:     true,
}

// CategoryUrgent 紧急公告：需要住户主动确认已读。
const CategoryUrgent = "紧急"
