package f

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



func TestF1(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	repo := repository.NewRecommendationRepository(db)
	if err := repo.Create(&model.Recommendation{StudentID: stu.ID, CounselorID: 1, UniversityIDs: "[1]"}); err != nil {
		t.Fatalf("create rec: %v", err)
	}
	if err := repo.Create(&model.Recommendation{StudentID: stu.ID, CounselorID: 1, UniversityIDs: "[2]"}); err != nil {
		t.Fatalf("create rec: %v", err)
	}
	items, err := repo.ListByStudent(stu.ID)
	if err != nil {
		t.Fatalf("ListByStudent() error = %v", err)
	}
	if len(items) != 2 || items[0].ID < items[1].ID {
		t.Fatalf("ListByStudent() order = %+v, want latest first", items)
	}
}


func TestF2(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	svc := newRecommendationService(db)
	rec, err := svc.Create(1, stu.ID, []uint{1, 2}, "reason")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if rec.UniversityIDs != "[1,2]" {
		t.Fatalf("UniversityIDs = %q, want %q", rec.UniversityIDs, "[1,2]")
	}
}


func TestF3(t *testing.T) {
	db := newTestDB(t)
	u1 := &model.University{Name: "U1", Country: "US", Ranking: 1}
	u2 := &model.University{Name: "U2", Country: "US", Ranking: 2}
	if err := db.Create(u1).Error; err != nil { t.Fatalf("create univ: %v", err) }
	if err := db.Create(u2).Error; err != nil { t.Fatalf("create univ: %v", err) }
	svc := newRecommendationService(db)
	rec := &model.Recommendation{StudentID: 1, CounselorID: 1, UniversityIDs: "[1,2]"}
	items, err := svc.ResolveUniversities(rec)
	if err != nil {
		t.Fatalf("ResolveUniversities() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ResolveUniversities() = %d, want 2", len(items))
	}
}


func TestF4(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	svc := newRecommendationService(db)
	rec, err := svc.Create(7, stu.ID, []uint{1}, "reason")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if rec.CounselorID != 7 {
		t.Fatalf("CounselorID = %d, want 7", rec.CounselorID)
	}
}


func TestF5(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	svc := newRecommendationService(db)
	_, err := svc.Create(1, stu.ID, nil, "reason")
	if err == nil {
		t.Fatalf("Create() with empty university IDs should return error")
	}
}
