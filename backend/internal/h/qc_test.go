package h

import (
	"log/slog"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/gbstudyapply/gbstudyapply/internal/config"
	"github.com/gbstudyapply/gbstudyapply/internal/constants"
	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
	. "github.com/gbstudyapply/gbstudyapply/internal/service"
)


func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.University{}, &model.ApplicationProject{}, &model.Document{},
		&model.DocumentVersion{}, &model.Annotation{}, &model.MaterialItem{}, &model.Message{},
		&model.Recommendation{}, &model.TimelineNode{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func newApplicationService(db *gorm.DB) *ApplicationService {
	return NewApplicationService(repository.NewApplicationProjectRepository(db), repository.NewUniversityRepository(db), slog.Default())
}

func newDocumentRepos(db *gorm.DB) (*repository.DocumentRepository, *repository.DocumentVersionRepository) {
	return repository.NewDocumentRepository(db), repository.NewDocumentVersionRepository(db)
}

func newRecommendationService(db *gorm.DB) *RecommendationService {
	return NewRecommendationService(repository.NewRecommendationRepository(db), repository.NewUniversityRepository(db), slog.Default())
}

func newUserService(t *testing.T, db *gorm.DB) *UserService {
	t.Helper()
	return NewUserService(repository.NewUserRepository(db), slog.Default(), &config.Config{JWTSecret: "secret", JWTExpire: time.Hour})
}

func newUniversityService(db *gorm.DB) *UniversityService {
	return NewUniversityService(repository.NewUniversityRepository(db), slog.Default())
}

func createUser(t *testing.T, db *gorm.DB, username, role string) *model.User {
	t.Helper()
	u := &model.User{Username: username, Email: username + "@test.com", PasswordHash: "x", Role: role}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func createUniversity(t *testing.T, db *gorm.DB) *model.University {
	t.Helper()
	u := &model.University{Name: "Test University", Country: "US", Ranking: 1}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("create university: %v", err)
	}
	return u
}

func createApplication(t *testing.T, db *gorm.DB, studentID, universityID uint) *model.ApplicationProject {
	t.Helper()
	a := &model.ApplicationProject{StudentID: studentID, UniversityID: universityID, Major: "CS", Status: constants.AppStatusPlanning}
	if err := db.Create(a).Error; err != nil {
		t.Fatalf("create application: %v", err)
	}
	return a
}

func createDocument(t *testing.T, db *gorm.DB, appID uint, title string) *model.Document {
	t.Helper()
	d := &model.Document{ApplicationID: appID, DocType: "ps", Title: title, Content: "v1", CurrentVersion: 1}
	if err := db.Create(d).Error; err != nil {
		t.Fatalf("create document: %v", err)
	}
	if err := db.Create(&model.DocumentVersion{DocumentID: d.ID, Content: "v1", VersionNo: 1, ChangeSummary: "init"}).Error; err != nil {
		t.Fatalf("create version: %v", err)
	}
	return d
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}



func TestH1(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	if _, err := svc.Create(&model.University{Name: "US Uni", Country: "US", Ranking: 1}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	if _, err := svc.Create(&model.University{Name: "UK Uni", Country: "UK", Ranking: 2}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	items, total, err := svc.List("US", 0, 0, "", 1, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Country != "US" {
		t.Fatalf("List(country=US) = total=%d items=%+v, want exactly one US university", total, items)
	}
}


func TestH2(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	if _, err := svc.Create(&model.University{Name: "Low", Country: "US", Ranking: 1}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	if _, err := svc.Create(&model.University{Name: "High", Country: "US", Ranking: 10}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	items, total, err := svc.List("", 5, 20, "", 1, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Ranking != 10 {
		t.Fatalf("List(ranking 5-20) = total=%d items=%+v, want one university with ranking 10", total, items)
	}
}


func TestH3(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	u, err := svc.Create(&model.University{Name: "NoJSON", Country: "US", Ranking: 1})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if u.TopMajors != "[]" || u.Requirements != "{}" {
		t.Fatalf("TopMajors=%q Requirements=%q, want [] and {}", u.TopMajors, u.Requirements)
	}
}


func TestH4(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	created, _ := svc.Create(&model.University{Name: "Old", Country: "US", Ranking: 1})
	updated, err := svc.Update(created.ID, &model.University{Name: "New", TuitionRange: "10-20", Requirements: "{\"gpa\":3.0}"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.TuitionRange != "10-20" {
		t.Fatalf("TuitionRange = %q, want %q", updated.TuitionRange, "10-20")
	}
}
