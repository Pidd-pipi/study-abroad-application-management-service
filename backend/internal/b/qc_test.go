package b

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



func TestB1(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	docRepo, verRepo := newDocumentRepos(db)
	svc := NewDocumentService(db, docRepo, verRepo, repository.NewAnnotationRepository(db), repository.NewApplicationProjectRepository(db), slog.Default())
	created, err := svc.Create(app.ID, "ps", "ps", "v1")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	versions, err := verRepo.ListByDocument(created.ID)
	if err != nil {
		t.Fatalf("ListByDocument() error = %v", err)
	}
	if len(versions) != 1 || versions[0].VersionNo != 1 {
		t.Fatalf("initial versions = %+v, want exactly one version with VersionNo=1", versions)
	}
}


func TestB2(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	d := createDocument(t, db, app.ID, "ps")
	docRepo, verRepo := newDocumentRepos(db)
	svc := NewDocumentService(db, docRepo, verRepo, repository.NewAnnotationRepository(db), repository.NewApplicationProjectRepository(db), slog.Default())
	saved, err := svc.Save(d.ID, "v2", "second")
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if saved.CurrentVersion != 2 {
		t.Fatalf("CurrentVersion = %d, want 2", saved.CurrentVersion)
	}
	versions, _ := verRepo.ListByDocument(d.ID)
	if len(versions) != 2 || versions[0].VersionNo != 2 {
		t.Fatalf("versions = %+v, want latest VersionNo=2", versions)
	}
}


func TestB3(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	d := createDocument(t, db, app.ID, "ps")
	_, verRepo := newDocumentRepos(db)
	if err := verRepo.Create(&model.DocumentVersion{DocumentID: d.ID, Content: "v2", VersionNo: 2}); err != nil {
		t.Fatalf("create v2: %v", err)
	}
	versions, _ := verRepo.ListByDocument(d.ID)
	if len(versions) < 2 || versions[0].VersionNo != 2 {
		t.Fatalf("ListByDocument order = %+v, want latest first", versions)
	}
}


func TestB4(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	d := createDocument(t, db, app.ID, "ps")
	docRepo, verRepo := newDocumentRepos(db)
	svc := NewDocumentService(db, docRepo, verRepo, repository.NewAnnotationRepository(db), repository.NewApplicationProjectRepository(db), slog.Default())
	if _, err := svc.Save(d.ID, "v2", "second"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	rolled, err := svc.Rollback(d.ID, 1)
	if err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}
	if rolled.CurrentVersion != 1 || rolled.Content != "v1" {
		t.Fatalf("Rollback() = version=%d content=%q, want version=1 content=v1", rolled.CurrentVersion, rolled.Content)
	}
}


func TestB5(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	docRepo, verRepo := newDocumentRepos(db)
	svc := NewDocumentService(db, docRepo, verRepo, repository.NewAnnotationRepository(db), repository.NewApplicationProjectRepository(db), slog.Default())
	_, err := svc.Create(app.ID, "ps", "", "content")
	if err == nil {
		t.Fatalf("Create() with empty title should return error")
	}
}


func TestB6(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	d := createDocument(t, db, app.ID, "ps")
	docRepo, verRepo := newDocumentRepos(db)
	svc := NewDocumentService(db, docRepo, verRepo, repository.NewAnnotationRepository(db), repository.NewApplicationProjectRepository(db), slog.Default())
	_, err := svc.Save(d.ID, "v2", "")
	if err == nil {
		t.Fatalf("Save() with empty summary should return error")
	}
}
