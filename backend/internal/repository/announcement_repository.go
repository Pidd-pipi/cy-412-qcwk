package repository

import (
	"errors"
	"time"

	"github.com/smartestate/smartestate/internal/model"
	"gorm.io/gorm"
)

type AnnouncementRepository struct{ DB *gorm.DB }

func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db}
}
func (r *AnnouncementRepository) Create(v *model.Announcement) error { return r.DB.Create(v).Error }

// List 物业/管理员看到全部公告。
func (r *AnnouncementRepository) List() (out []model.Announcement, e error) {
	e = r.DB.Preload("Publisher").Order("top desc, publish_at desc").Find(&out).Error
	return
}

// ListForResident 住户只看到与自己房产匹配的公告：
// 全小区、同楼栋、同楼栋同单元。
func (r *AnnouncementRepository) ListForResident(building, unit string) (out []model.Announcement, e error) {
	q := r.DB.Preload("Publisher").Order("top desc, publish_at desc")
	q = q.Where("scope = ?", "all").
		Or("scope = ? AND building = ?", "building", building).
		Or("scope = ? AND building = ? AND unit = ?", "unit", building, unit)
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

// VisibleByResident 判断某条公告是否对指定房产可见。
func (r *AnnouncementRepository) VisibleByResident(v model.Announcement, building, unit string) bool {
	switch v.Scope {
	case "building":
		return v.Building != "" && v.Building == building
	case "unit":
		return v.Building != "" && v.Building == building && v.Unit != "" && v.Unit == unit
	default: // all
		return true
	}
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

// ConfirmRead 住户确认紧急公告已读。重复确认不会重复计数：
// 唯一索引 + 已确认状态判断保证幂等。
func (r *AnnouncementRepository) ConfirmRead(id, uid uint) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		var rcd model.AnnouncementRead
		e := tx.Where("announcement_id=? AND user_id=?", id, uid).First(&rcd).Error
		if errors.Is(e, gorm.ErrRecordNotFound) {
			now := time.Now()
			return tx.Create(&model.AnnouncementRead{AnnouncementID: id, UserID: uid, Confirmed: true, ConfirmedAt: &now, CreatedAt: now}).Error
		}
		if e != nil {
			return e
		}
		if rcd.Confirmed {
			return nil
		}
		now := time.Now()
		return tx.Model(&rcd).Updates(map[string]any{"confirmed": true, "confirmed_at": &now}).Error
	})
}

// IsConfirmed 当前住户是否已确认某条紧急公告。
func (r *AnnouncementRepository) IsConfirmed(id, uid uint) bool {
	var n int64
	r.DB.Model(&model.AnnouncementRead{}).
		Where("announcement_id=? AND user_id=? AND confirmed=?", id, uid, true).Count(&n)
	return n > 0
}

// TargetResidents 公告影响范围内的住户（需绑定房产）。
func (r *AnnouncementRepository) TargetResidents(v model.Announcement) (out []model.User, e error) {
	q := r.DB.Where("role=?", "resident").
		Where("building <> ''")
	switch v.Scope {
	case "building":
		q = q.Where("building = ?", v.Building)
	case "unit":
		q = q.Where("building = ? AND unit = ?", v.Building, v.Unit)
	}
	e = q.Order("building, unit, room").Find(&out).Error
	return
}

// UnconfirmedResidents 返回影响范围内尚未确认的住户。
func (r *AnnouncementRepository) UnconfirmedResidents(v model.Announcement) (out []model.User, e error) {
	target, e := r.TargetResidents(v)
	if e != nil {
		return
	}
	confirmed := map[uint]bool{}
	var rows []model.AnnouncementRead
	if e = r.DB.Where("announcement_id=? AND confirmed=?", v.ID, true).Find(&rows).Error; e != nil {
		return
	}
	for _, x := range rows {
		confirmed[x.UserID] = true
	}
	for _, u := range target {
		if !confirmed[u.ID] {
			out = append(out, u)
		}
	}
	return
}
