package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/dto"
	"github.com/smartestate/smartestate/internal/service"
)

type AnnouncementHandler struct {
	Handler
	svc *service.AnnouncementService
}

func NewAnnouncementHandler(s *service.AnnouncementService, h *Handler) *AnnouncementHandler {
	return &AnnouncementHandler{Handler: *h, svc: s}
}
func (h *AnnouncementHandler) List(c *gin.Context) {
	v, e := h.svc.List(c.GetUint("userID"), c.GetString("role"))
	if e != nil {
		Fail(c, 500, constants.CodeInternal, e.Error())
		return
	}
	OK(c, v)
}
func (h *AnnouncementHandler) Create(c *gin.Context) {
	var r dto.CreateAnnouncementRequest
	if !Bind(c, &r, h.Validate) {
		return
	}
	v, e := h.svc.Create(c.GetUint("userID"), r.Title, r.Content, r.Category, r.Scope, r.Building, r.Unit, r.Top)
	if e != nil {
		Fail(c, 400, constants.CodeBadRequest, e.Error())
		return
	}
	OK(c, v)
}
func (h *AnnouncementHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.Detail(uint(id), c.GetUint("userID"), c.GetString("role"))
	if e != nil {
		Fail(c, 404, constants.CodeNotFound, e.Error())
		return
	}
	OK(c, v)
}

// Confirm 住户确认紧急公告已读，幂等。
func (h *AnnouncementHandler) Confirm(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.Confirm(uint(id), c.GetUint("userID"))
	if e != nil {
		Fail(c, 400, constants.CodeBadRequest, e.Error())
		return
	}
	OK(c, v)
}

// ScopeOptions 发布公告时可选的楼栋 / 单元，仅发布权限可用。
func (h *AnnouncementHandler) ScopeOptions(c *gin.Context) {
	buildings, units, e := h.svc.ScopeOptions(c.Query("building"))
	if e != nil {
		Fail(c, 500, constants.CodeInternal, e.Error())
		return
	}
	OK(c, gin.H{"buildings": buildings, "units": units})
}
