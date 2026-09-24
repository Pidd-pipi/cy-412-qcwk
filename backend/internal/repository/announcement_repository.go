package repository

import (
	"errors"
	"github.com/smartestate/smartestate/internal/model"
	"gorm.io/gorm"
)

type AnnouncementRepository struct{ DB *gorm.DB }

func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db}
}
func (r *AnnouncementRepository) Create(v *model.Announcement) error { return r.DB.Create(v).Error }

// ScopeMatch builds the WHERE clause that limits announcements to a
// resident's property: 全小区, or same building / same building+unit.
func ScopeMatch(building, unit string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("scope = ?", model.AnnouncementScopeAll).
			Or("scope = ? AND scope_building = ?", model.AnnouncementScopeBuilding, building).
			Or("scope = ? AND scope_building = ? AND scope_unit = ?", model.AnnouncementScopeUnit, building, unit)
	}
}

func (r *AnnouncementRepository) List(scope func(*gorm.DB) *gorm.DB) (out []model.Announcement, e error) {
	q := r.DB.Preload("Publisher").Order("top desc, publish_at desc")
	if scope != nil {
		q = q.Scopes(scope)
	}
	e = q.Find(&out).Error
	return
}
func (r *AnnouncementRepository) ByID(id uint) (v model.Announcement, e error) {
	e = r.DB.Preload("Publisher").First(&v, id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		e = ErrNotFound
	}
	return
}
func (r *AnnouncementRepository) MarkRead(id, uid uint) error {
	var rcd model.AnnouncementRead
	if e := r.DB.Where("announcement_id=? AND user_id=?", id, uid).First(&rcd).Error; errors.Is(e, gorm.ErrRecordNotFound) {
		if e = r.DB.Create(&model.AnnouncementRead{AnnouncementID: id, UserID: uid}).Error; e != nil {
			return e
		}
		return r.DB.Model(&model.Announcement{}).Where("id=?", id).UpdateColumn("read_count", gorm.Expr("read_count + 1")).Error
	}
	return nil
}

// HasConfirm reports whether the user already confirmed an announcement.
func (r *AnnouncementRepository) HasConfirm(id, uid uint) (bool, error) {
	var n int64
	e := r.DB.Model(&model.AnnouncementConfirm{}).Where("announcement_id=? AND user_id=?", id, uid).Count(&n).Error
	return n > 0, e
}

// Confirm records a confirmation idempotently; created is false when the
// user had already confirmed, so repeated calls never inflate the count.
func (r *AnnouncementRepository) Confirm(id, uid uint) (bool, error) {
	var rcd model.AnnouncementConfirm
	e := r.DB.Where("announcement_id=? AND user_id=?", id, uid).First(&rcd).Error
	if e == nil {
		return false, nil
	}
	if !errors.Is(e, gorm.ErrRecordNotFound) {
		return false, e
	}
	if e = r.DB.Create(&model.AnnouncementConfirm{AnnouncementID: id, UserID: uid}).Error; e != nil {
		return false, e
	}
	return true, r.DB.Model(&model.Announcement{}).Where("id=?", id).
		UpdateColumn("confirmed_count", gorm.Expr("confirmed_count + 1")).Error
}

// TargetResidents returns residents covered by an announcement's scope.
func (r *AnnouncementRepository) TargetResidents(a model.Announcement) (out []model.ShortUser, e error) {
	q := r.DB.Model(&model.User{}).Where("role = ?", "resident")
	switch a.Scope {
	case model.AnnouncementScopeBuilding:
		q = q.Where("building = ?", a.ScopeBuilding)
	case model.AnnouncementScopeUnit:
		q = q.Where("building = ? AND unit = ?", a.ScopeBuilding, a.ScopeUnit)
	}
	e = q.Order("building, unit, room").Find(&out).Error
	return
}

// UnconfirmedResidents returns residents in scope who have not confirmed.
func (r *AnnouncementRepository) UnconfirmedResidents(a model.Announcement) (out []model.ShortUser, e error) {
	sub := r.DB.Model(&model.AnnouncementConfirm{}).Select("user_id").Where("announcement_id = ?", a.ID)
	q := r.DB.Model(&model.User{}).Where("role = ? AND id NOT IN (?)", "resident", sub)
	switch a.Scope {
	case model.AnnouncementScopeBuilding:
		q = q.Where("building = ?", a.ScopeBuilding)
	case model.AnnouncementScopeUnit:
		q = q.Where("building = ? AND unit = ?", a.ScopeBuilding, a.ScopeUnit)
	}
	e = q.Order("building, unit, room").Find(&out).Error
	return
}
