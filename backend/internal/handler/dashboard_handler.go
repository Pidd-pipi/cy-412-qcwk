package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/smartestate/smartestate/internal/service"
)

type DashboardHandler struct {
	repairs  *service.RepairService
	payments *service.PaymentService
	anns     *service.AnnouncementService
}

func NewDashboardHandler(r *service.RepairService, p *service.PaymentService, a *service.AnnouncementService) *DashboardHandler {
	return &DashboardHandler{r, p, a}
}
func (h *DashboardHandler) Summary(c *gin.Context) {
	open, _ := h.repairs.OpenCount()
	amount, _ := h.payments.MonthlyPaid()
	// 公告同样按当前用户角色与房产过滤，住户首页只出现与自己房产匹配的公告。
	anns, _ := h.anns.List(c.GetUint("userID"), c.GetString("role"))
	if len(anns) > 3 {
		anns = anns[:3]
	}
	OK(c, gin.H{"pending_repairs": open, "monthly_paid": amount, "announcements": anns})
}
