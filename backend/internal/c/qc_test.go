package c

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



func TestC1(t *testing.T) {
	if !containsString(constants.ValidMaterialStatuses(), constants.MaterialApproved) {
		t.Fatalf("ValidMaterialStatuses() = %v, want contain %q", constants.ValidMaterialStatuses(), constants.MaterialApproved)
	}
}


func TestC2(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewMaterialItemRepository(db)
	svc := NewMaterialService(repo, slog.Default())
	for _, item := range []model.MaterialItem{
		{ApplicationID: app.ID, Name: "a", Status: constants.MaterialPending},
		{ApplicationID: app.ID, Name: "b", Status: constants.MaterialUploaded},
		{ApplicationID: app.ID, Name: "c", Status: constants.MaterialApproved},
	} {
		if err := repo.Create(&item); err != nil {
			t.Fatalf("create material: %v", err)
		}
	}
	got, err := svc.Progress(app.ID)
	if err != nil {
		t.Fatalf("Progress() error = %v", err)
	}
	if got != 66 {
		t.Fatalf("Progress() = %d, want 66", got)
	}
}


func TestC3(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewMaterialItemRepository(db)
	svc := NewMaterialService(repo, slog.Default())
	item := &model.MaterialItem{ApplicationID: app.ID, Name: "a", Status: constants.MaterialPending}
	if err := repo.Create(item); err != nil {
		t.Fatalf("create material: %v", err)
	}
	updated, err := svc.UpdateStatus(1, item.ID, constants.RoleStudent, constants.MaterialUploaded, "http://file")
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if updated.UploadedAt == nil {
		t.Fatalf("UploadedAt = nil, want non-nil")
	}
}


func TestC4(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewMaterialItemRepository(db)
	for i := 0; i < 3; i++ {
		if err := repo.Create(&model.MaterialItem{ApplicationID: app.ID, Name: "m", Status: constants.MaterialPending}); err != nil {
			t.Fatalf("create material: %v", err)
		}
	}
	items, err := repo.ListByApplication(app.ID)
	if err != nil {
		t.Fatalf("ListByApplication() error = %v", err)
	}
	if len(items) != 3 || items[0].ID > items[1].ID {
		t.Fatalf("ListByApplication() order = %+v, want ascending id", items)
	}
}


func TestC5(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewMaterialItemRepository(db)
	svc := NewMaterialService(repo, slog.Default())
	item := &model.MaterialItem{ApplicationID: app.ID, Name: "a", Status: constants.MaterialPending}
	if err := repo.Create(item); err != nil {
		t.Fatalf("create material: %v", err)
	}
	_, err := svc.UpdateStatus(1, item.ID, constants.RoleStudent, constants.MaterialUploaded, "")
	if err == nil {
		t.Fatalf("UpdateStatus(uploaded, empty url) should return error")
	}
}
