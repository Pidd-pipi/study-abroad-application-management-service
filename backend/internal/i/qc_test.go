package i

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



func TestI1(t *testing.T) {
	db := newTestDB(t)
	repo := repository.NewUserRepository(db)
	if err := repo.Create(&model.User{Username: "stu1", Email: "stu1@test.com", PasswordHash: "x", Role: constants.RoleStudent}); err != nil {
		t.Fatalf("create student: %v", err)
	}
	if err := repo.Create(&model.User{Username: "coun1", Email: "coun1@test.com", PasswordHash: "x", Role: constants.RoleCounselor}); err != nil {
		t.Fatalf("create counselor: %v", err)
	}
	items, err := repo.ListStudents()
	if err != nil {
		t.Fatalf("ListStudents() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ListStudents() = %d, want 1", len(items))
	}
}


func TestI2(t *testing.T) {
	db := newTestDB(t)
	svc := newUserService(t, db)
	if _, _, err := svc.Register("stu1", "stu1@test.com", "secret", "", ""); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	_, _, err := svc.Login("stu1", "wrong")
	if err == nil {
		t.Fatalf("Login() with wrong password should return error")
	}
}


func TestI3(t *testing.T) {
	db := newTestDB(t)
	svc := newUserService(t, db)
	u, _, err := svc.Register("stu1", "stu1@test.com", "secret", "", "")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	updated, err := svc.UpdateProfile(u.ID, "", "13800000000")
	if err != nil {
		t.Fatalf("UpdateProfile() error = %v", err)
	}
	if updated.Phone != "13800000000" {
		t.Fatalf("Phone = %q, want %q", updated.Phone, "13800000000")
	}
}


func TestI4(t *testing.T) {
	db := newTestDB(t)
	svc := newUserService(t, db)
	u, _, err := svc.Register("stu1", "stu1@test.com", "secret", "", "")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if u.Role != constants.RoleStudent {
		t.Fatalf("Role = %q, want %q", u.Role, constants.RoleStudent)
	}
}


func TestI5(t *testing.T) {
	db := newTestDB(t)
	svc := newUserService(t, db)
	_, _, err := svc.Register("", "stu@test.com", "secret", "", "")
	if err == nil {
		t.Fatalf("Register() with empty username should return error")
	}
}


func TestI6(t *testing.T) {
	db := newTestDB(t)
	svc := newUserService(t, db)
	_, _, err := svc.Register("stu1", "", "secret", "", "")
	if err == nil {
		t.Fatalf("Register() with empty email should return error")
	}
}
