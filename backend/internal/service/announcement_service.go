package service

import (
	"fmt"
	"github.com/smartestate/smartestate/internal/constants"
	"github.com/smartestate/smartestate/internal/model"
	"github.com/smartestate/smartestate/internal/repository"
	"log/slog"
	"time"
)

type AnnouncementService struct {
	repo   *repository.AnnouncementRepository
	users  *repository.UserRepository
	logger *slog.Logger
}

func NewAnnouncementService(r *repository.AnnouncementRepository, u *repository.UserRepository, l *slog.Logger) *AnnouncementService {
	return &AnnouncementService{r, u, l}
}

// Create 发布公告。scope=all 时 building/unit 留空；
// building 必须带楼栋；unit 必须带楼栋与单元。
func (s *AnnouncementService) Create(uid uint, title, content, category, scope, building, unit string, top bool) (model.Announcement, error) {
	if !constants.ValidAnnouncementScopes[scope] {
		return model.Announcement{}, fmt.Errorf("Announcement create failed: invalid scope=%s, publisher=%d", scope, uid)
	}
	switch scope {
	case constants.AnnouncementScopeBuilding:
		if building == "" {
			return model.Announcement{}, fmt.Errorf("Announcement create failed: building scope requires building, publisher=%d", uid)
		}
		unit = ""
	case constants.AnnouncementScopeUnit:
		if building == "" || unit == "" {
			return model.Announcement{}, fmt.Errorf("Announcement create failed: unit scope requires building and unit, publisher=%d", uid)
		}
	case constants.AnnouncementScopeAll:
		building, unit = "", ""
	}
	v := model.Announcement{
		Title: title, Content: content, Category: category,
		PublisherID: uid, Top: top, PublishAt: time.Now(),
		Scope: scope, Building: building, Unit: unit,
	}
	if e := s.repo.Create(&v); e != nil {
		return v, fmt.Errorf("Announcement[scope=%s] create failed: %w", scope, e)
	}
	return v, nil
}

// List 列表：住户按房产过滤；物业/管理员看到全部并附带紧急公告统计。
func (s *AnnouncementService) List(uid uint, role string) ([]model.Announcement, error) {
	if role == constants.UserRoleResident {
		viewer, e := s.users.ByID(uid)
		if e != nil {
			return nil, fmt.Errorf("Announcement list failed: resident=%d not found: %w", uid, e)
		}
		out, e := s.repo.ListForResident(viewer.Building, viewer.Unit)
		if e != nil {
			return out, fmt.Errorf("Announcement list failed for resident=%d: %w", uid, e)
		}
		for i := range out {
			if out[i].Category == constants.CategoryUrgent {
				out[i].Confirmed = s.repo.IsConfirmed(out[i].ID, uid)
			}
		}
		return out, nil
	}
	out, e := s.repo.List()
	if e != nil {
		return out, fmt.Errorf("Announcement list failed for role=%s: %w", role, e)
	}
	for i := range out {
		if out[i].Category == constants.CategoryUrgent {
			s.fillUrgentStats(&out[i])
		}
	}
	return out, nil
}

// Detail 详情：普通公告记录阅读并回写 read_count；
// 紧急公告不自动计数，住户需显式确认。
func (s *AnnouncementService) Detail(id, uid uint, role string) (model.Announcement, error) {
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	viewer, e := s.users.ByID(uid)
	if e != nil {
		return v, fmt.Errorf("Announcement[id=%d] detail failed: viewer=%d not found, current role=%s: %w", id, uid, role, e)
	}
	if role == constants.UserRoleResident && !s.repo.VisibleByResident(v, viewer.Building, viewer.Unit) {
		return v, fmt.Errorf("Announcement[id=%d] detail forbidden: outside scope, resident=%d, current role=%s", id, uid, role)
	}

	if role == constants.UserRoleResident {
		if v.Category == constants.CategoryUrgent {
			v.Confirmed = s.repo.IsConfirmed(id, uid)
		} else {
			if e = s.repo.MarkRead(id, uid); e != nil {
				return v, fmt.Errorf("Announcement[id=%d] mark read failed: %w", id, e)
			}
			v, e = s.repo.ByID(id)
			if e != nil {
				return v, e
			}
		}
		return v, nil
	}

	if v.Category == constants.CategoryUrgent {
		s.fillUrgentStats(&v)
	}
	return v, nil
}

// Confirm 住户确认紧急公告已读，幂等：重复确认不增加人数。
// 仅影响范围内的住户可确认。
func (s *AnnouncementService) Confirm(id, uid uint) (model.Announcement, error) {
	v, e := s.repo.ByID(id)
	if e != nil {
		return v, e
	}
	if v.Category != constants.CategoryUrgent {
		return v, fmt.Errorf("Announcement[id=%d] confirm failed: not an urgent announcement", id)
	}
	viewer, e := s.users.ByID(uid)
	if e != nil {
		return v, fmt.Errorf("Announcement[id=%d] confirm failed: resident=%d not found: %w", id, uid, e)
	}
	if !s.repo.VisibleByResident(v, viewer.Building, viewer.Unit) {
		return v, fmt.Errorf("Announcement[id=%d] confirm forbidden: outside scope, resident=%d", id, uid)
	}
	if e = s.repo.ConfirmRead(id, uid); e != nil {
		return v, fmt.Errorf("Announcement[id=%d] confirm failed for resident=%d: %w", id, uid, e)
	}
	v.Confirmed = true
	return v, nil
}

// ScopeOptions 返回发布公告可选的楼栋，以及某楼栋下可选的单元。
// 取值来源于住户已绑定房产，保证范围与房产一致。
func (s *AnnouncementService) ScopeOptions(building string) ([]string, []string, error) {
	buildings, e := s.users.Buildings()
	if e != nil {
		return nil, nil, fmt.Errorf("Announcement scope options buildings failed: %w", e)
	}
	if building == "" {
		return buildings, nil, nil
	}
	units, e := s.users.Units(building)
	if e != nil {
		return buildings, nil, fmt.Errorf("Announcement scope options units failed for building=%s: %w", building, e)
	}
	return buildings, units, nil
}

// fillUrgentStats 物业视角：填充确认人数与未确认名单。
func (s *AnnouncementService) fillUrgentStats(v *model.Announcement) {
	targets, e := s.repo.TargetResidents(*v)
	if e != nil {
		s.logger.Error("announcement urgent stats target query failed", "id", v.ID, "err", e)
		return
	}
	unconfirmed, e := s.repo.UnconfirmedResidents(*v)
	if e != nil {
		s.logger.Error("announcement urgent stats unconfirmed query failed", "id", v.ID, "err", e)
		return
	}
	v.ConfirmedCount = len(targets) - len(unconfirmed)
	v.UnconfirmedUsers = unconfirmed
}
