package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/smartestate/smartestate/internal/dto"
	"github.com/smartestate/smartestate/internal/repository"
	"github.com/smartestate/smartestate/internal/service"
	"strconv"
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
		Fail(c, 500, 50001, e.Error())
		return
	}
	OK(c, v)
}
func (h *AnnouncementHandler) Create(c *gin.Context) {
	var r dto.CreateAnnouncementRequest
	if !Bind(c, &r, h.Validate) {
		return
	}
	v, e := h.svc.Create(c.GetUint("userID"), c.GetString("role"), r.Title, r.Content, r.Category, r.Scope, r.ScopeBuilding, r.ScopeUnit, r.Top)
	if e != nil {
		if errors.Is(e, service.ErrBadScope) {
			Fail(c, 400, 40001, e.Error())
			return
		}
		Fail(c, 500, 50001, e.Error())
		return
	}
	OK(c, v)
}
func (h *AnnouncementHandler) Detail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.Detail(uint(id), c.GetUint("userID"), c.GetString("role"))
	if e != nil {
		Fail(c, 404, 40401, e.Error())
		return
	}
	OK(c, v)
}
func (h *AnnouncementHandler) Confirm(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	v, e := h.svc.Confirm(uint(id), c.GetUint("userID"))
	if e != nil {
		if errors.Is(e, repository.ErrNotFound) {
			Fail(c, 404, 40401, e.Error())
			return
		}
		Fail(c, 500, 50001, e.Error())
		return
	}
	OK(c, v)
}
