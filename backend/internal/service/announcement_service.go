package service

import (
	"errors"
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"gorm.io/gorm"
	"log/slog"
	"time"
)

// ErrBadScope indicates a scoped announcement is missing building/unit.
var ErrBadScope = errors.New("announcement scope requires building and unit")

type AnnouncementService struct {
	repo   *repository.AnnouncementRepository
	users  *repository.UserRepository
	logger *slog.Logger
}

func NewAnnouncementService(r *repository.AnnouncementRepository, u *repository.UserRepository, l *slog.Logger) *AnnouncementService {
	return &AnnouncementService{repo: r, users: u, logger: l}
}

func isStaff(role string) bool {
	return role == constants.UserRoleStaff || role == constants.UserRoleAdmin
}

// inScope reports whether an announcement targets a resident's property.
func inScope(a model.Announcement, building, unit string) bool {
	switch a.Scope {
	case model.AnnouncementScopeAll:
		return true
	case model.AnnouncementScopeBuilding:
		return a.ScopeBuilding == building
	case model.AnnouncementScopeUnit:
		return a.ScopeBuilding == building && a.ScopeUnit == unit
	}
	return true
}

func isUrgent(a model.Announcement) bool { return a.Category == "紧急" }

func (s *AnnouncementService) Create(uid uint, role string, title, content, category, scope, building, unit string, top bool) (model.Announcement, error) {
	if scope == "" {
		scope = model.AnnouncementScopeAll
	}
	if scope != model.AnnouncementScopeAll && (building == "" || (scope == model.AnnouncementScopeUnit && unit == "")) {
		return model.Announcement{}, ErrBadScope
	}
	if scope == model.AnnouncementScopeBuilding {
		unit = ""
	}
	if scope == model.AnnouncementScopeAll {
		building, unit = "", ""
	}
	v := model.Announcement{Title: title, Content: content, Category: category, Scope: scope, ScopeBuilding: building, ScopeUnit: unit, PublisherID: uid, Top: top, PublishAt: time.Now()}
	e := s.repo.Create(&v)
	return v, e
}

func (s *AnnouncementService) List(uid uint, role string) ([]model.Announcement, error) {
	u, e := s.users.ByID(uid)
	if e != nil {
		return nil, e
	}
	var scope func(*gorm.DB) *gorm.DB
	if !isStaff(role) {
		scope = repository.ScopeMatch(u.Building, u.Unit)
	}
	items, e := s.repo.List(scope)
	if e != nil {
		return nil, e
	}
	if isStaff(role) {
		for i := range items {
			if isUrgent(items[i]) {
				items[i].UnconfirmedUsers, e = s.repo.UnconfirmedResidents(items[i])
				if e != nil {
					return nil, e
				}
			}
		}
		return items, nil
	}
	// Attach each resident's confirmation state for urgent announcements.
	for i := range items {
		if isUrgent(items[i]) {
			items[i].Confirmed, e = s.repo.HasConfirm(items[i].ID, uid)
			if e != nil {
				return nil, e
			}
		}
	}
	return items, nil
}

func (s *AnnouncementService) Detail(id, uid uint, role string) (model.Announcement, error) {
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if !isStaff(role) {
		u, e := s.users.ByID(uid)
		if e != nil {
			return v, e
		}
		if !inScope(v, u.Building, u.Unit) {
			return v, repository.ErrNotFound
		}
	}
	if e = s.repo.MarkRead(id, uid); e != nil {
		return v, e
	}
	// Reload so the response immediately contains the incremented read count.
	if v, e = s.repo.ByID(id); e != nil {
		return v, e
	}
	if isStaff(role) {
		if isUrgent(v) {
			v.UnconfirmedUsers, e = s.repo.UnconfirmedResidents(v)
		}
		return v, e
	}
	if isUrgent(v) {
		v.Confirmed, e = s.repo.HasConfirm(id, uid)
	}
	return v, e
}

// Confirm records a resident's explicit 已读确认 for an urgent, in-scope
// announcement. Repeated confirmations are ignored (count never changes).
func (s *AnnouncementService) Confirm(id, uid uint) (model.Announcement, error) {
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	u, e := s.users.ByID(uid)
	if e != nil {
		return v, e
	}
	if !isUrgent(v) || !inScope(v, u.Building, u.Unit) {
		return v, repository.ErrNotFound
	}
	if _, e = s.repo.Confirm(id, uid); e != nil {
		return v, e
	}
	v, e = s.repo.ByID(id)
	if e == nil {
		v.Confirmed = true
	}
	return v, e
}
