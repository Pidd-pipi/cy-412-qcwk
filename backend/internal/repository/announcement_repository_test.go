package repository

import (
	"github.com/smartestate/smartestate/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func newAnnouncementTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, e := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if e != nil {
		t.Fatal(e)
	}
	if e = db.AutoMigrate(&model.User{}, &model.Announcement{}, &model.AnnouncementRead{}, &model.AnnouncementConfirm{}); e != nil {
		t.Fatal(e)
	}
	return db
}

func TestAnnouncementListScopeMatch(t *testing.T) {
	db := newAnnouncementTestDB(t)
	pub := model.User{Phone: "staff", Nickname: "s", Role: "staff"}
	db.Create(&pub)
	db.Create(&model.Announcement{Title: "all", Category: "通知", Scope: model.AnnouncementScopeAll, PublisherID: pub.ID})
	db.Create(&model.Announcement{Title: "b1", Category: "通知", Scope: model.AnnouncementScopeBuilding, ScopeBuilding: "1栋", PublisherID: pub.ID})
	db.Create(&model.Announcement{Title: "u1-2", Category: "紧急", Scope: model.AnnouncementScopeUnit, ScopeBuilding: "1栋", ScopeUnit: "2单元", PublisherID: pub.ID})
	db.Create(&model.Announcement{Title: "b3", Category: "通知", Scope: model.AnnouncementScopeBuilding, ScopeBuilding: "3栋", PublisherID: pub.ID})

	r := NewAnnouncementRepository(db)
	got, e := r.List(ScopeMatch("1栋", "2单元"))
	if e != nil {
		t.Fatal(e)
	}
	var titles []string
	for _, a := range got {
		titles = append(titles, a.Title)
	}
	want := map[string]bool{"all": true, "b1": true, "u1-2": true}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %d in-scope announcements", titles, len(want))
	}
	for _, v := range titles {
		if !want[v] {
			t.Fatalf("out-of-scope announcement %q leaked", v)
		}
	}
}

func TestAnnouncementConfirmIdempotent(t *testing.T) {
	db := newAnnouncementTestDB(t)
	u := model.User{Phone: "r", Nickname: "张业主", Role: "resident", Building: "1栋", Unit: "2单元", Room: "802"}
	db.Create(&u)
	a := model.Announcement{Title: "停水", Category: "紧急", Scope: model.AnnouncementScopeUnit, ScopeBuilding: "1栋", ScopeUnit: "2单元"}
	db.Create(&a)
	r := NewAnnouncementRepository(db)

	created, e := r.Confirm(a.ID, u.ID)
	if e != nil || !created {
		t.Fatalf("first confirm: created=%v err=%v", created, e)
	}
	created, e = r.Confirm(a.ID, u.ID)
	if e != nil || created {
		t.Fatalf("second confirm must be idempotent: created=%v err=%v", created, e)
	}
	got, _ := r.ByID(a.ID)
	if got.ConfirmedCount != 1 {
		t.Fatalf("confirmed_count=%d, want 1", got.ConfirmedCount)
	}

	miss, _ := r.UnconfirmedResidents(got)
	if len(miss) != 0 {
		t.Fatalf("confirmed resident still listed as unconfirmed: %+v", miss)
	}
}

func TestAnnouncementUnconfirmedResidentsScoped(t *testing.T) {
	db := newAnnouncementTestDB(t)
	u1 := model.User{Phone: "1", Nickname: "u1", Role: "resident", Building: "1栋", Unit: "2单元", Room: "801"}
	u2 := model.User{Phone: "2", Nickname: "u2", Role: "resident", Building: "1栋", Unit: "2单元", Room: "802"}
	u3 := model.User{Phone: "3", Nickname: "u3", Role: "resident", Building: "1栋", Unit: "1单元", Room: "301"}
	db.Create(&u1)
	db.Create(&u2)
	db.Create(&u3)
	a := model.Announcement{Title: "停水", Category: "紧急", Scope: model.AnnouncementScopeUnit, ScopeBuilding: "1栋", ScopeUnit: "2单元"}
	db.Create(&a)
	r := NewAnnouncementRepository(db)

	miss, e := r.UnconfirmedResidents(a)
	if e != nil {
		t.Fatal(e)
	}
	if len(miss) != 2 {
		t.Fatalf("got %d unconfirmed, want 2 (u1,u2 only)", len(miss))
	}
	if _, e = r.Confirm(a.ID, u1.ID); e != nil {
		t.Fatal(e)
	}
	miss, _ = r.UnconfirmedResidents(a)
	if len(miss) != 1 || miss[0].ID != u2.ID {
		t.Fatalf("after confirm got %+v, want only u2", miss)
	}
}
