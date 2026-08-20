package j

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
	"github.com/gbstudyapply/gbstudyapply/internal/util"
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



func TestJ1(t *testing.T) {
	projects := []model.ApplicationProject{{Status: "planning"}, {Status: "submitted"}}
	s := ComputeAppStats(projects)
	if s.Total != 2 {
		t.Fatalf("Total = %d, want 2", s.Total)
	}
}


func TestJ2(t *testing.T) {
	projects := []model.ApplicationProject{{Status: "admitted"}}
	s := ComputeAppStats(projects)
	if s.Admitted != 1 {
		t.Fatalf("Admitted = %d, want 1", s.Admitted)
	}
}


func TestJ3(t *testing.T) {
	projects := []model.ApplicationProject{{Status: "waitlisted"}}
	s := ComputeAppStats(projects)
	if s.Applied != 1 {
		t.Fatalf("Applied = %d, want 1", s.Applied)
	}
}


func TestJ4(t *testing.T) {
	token, err := util.GenerateToken(1, "stu", "student", "secret", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}
	claims, err := util.ParseToken(token, "secret")
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.UserID != 1 {
		t.Fatalf("UserID = %d, want 1", claims.UserID)
	}
}


func TestJ5(t *testing.T) {
	_, err := util.GenerateToken(1, "stu", "student", "", time.Hour)
	if err == nil {
		t.Fatalf("GenerateToken() with empty secret should return error")
	}
}


func TestJ6(t *testing.T) {
	_, err := util.GenerateToken(1, "stu", "student", "secret", 0)
	if err == nil {
		t.Fatalf("GenerateToken() with non-positive expire should return error")
	}
}
