package router

import (
	"github.com/gin-gonic/gin"
	"github.com/smartestate/smartestate/internal/handler"
	"github.com/smartestate/smartestate/internal/middleware"
)

func RegisterAnnouncements(g *gin.RouterGroup, sv Services, h *handler.Handler) {
	x := handler.NewAnnouncementHandler(sv.Announcements, h)
	g.GET("/announcements", x.List)
	g.GET("/announcements/:id", x.Detail)
	// 住户对紧急公告确认已读，幂等。
	g.POST("/announcements/:id/confirm", middleware.OperationLog(sv.Logs, "announcement.confirm"), x.Confirm)
	g.GET("/announcement-scope-options", middleware.RequirePermission(sv.Permissions, "announcement:publish"), x.ScopeOptions)
	g.POST("/announcements", middleware.RequirePermission(sv.Permissions, "announcement:publish"), middleware.OperationLog(sv.Logs, "announcement.create"), x.Create)
}
