package repository

import (
	"errors"
	"github.com/smartestate/smartestate/internal/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")

type UserRepository struct{ DB *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db} }
func (r *UserRepository) ByID(id uint) (model.User, error) {
	var v model.User
	e := r.DB.First(&v, id).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return v, ErrNotFound
	}
	return v, e
}
func (r *UserRepository) ByPhone(phone string) (model.User, error) {
	var v model.User
	e := r.DB.Where("phone = ?", phone).First(&v).Error
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return v, ErrNotFound
	}
	return v, e
}
func (r *UserRepository) Update(v *model.User) error { return r.DB.Save(v).Error }
func (r *UserRepository) ListStaff() (out []model.User, e error) {
	e = r.DB.Where("role IN ?", []string{"staff", "admin"}).Find(&out).Error
	return
}

// Buildings 住户已绑定房产中出现过的楼栋（供发布公告选择范围）。
func (r *UserRepository) Buildings() (out []string, e error) {
	e = r.DB.Model(&model.User{}).
		Where("role = ? AND building <> ''", "resident").
		Distinct("building").Order("building").Pluck("building", &out).Error
	return
}

// Units 指定楼栋下住户已绑定房产中出现过的单元。
func (r *UserRepository) Units(building string) (out []string, e error) {
	e = r.DB.Model(&model.User{}).
		Where("role = ? AND building = ? AND unit <> ''", "resident", building).
		Distinct("unit").Order("unit").Pluck("unit", &out).Error
	return
}
