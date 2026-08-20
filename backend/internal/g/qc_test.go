package g

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



func TestG1(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	if err := repo.Create(&model.TimelineNode{ApplicationID: app.ID, Title: "due", DueDate: time.Now().AddDate(0,0,1)}); err != nil {
		t.Fatalf("create node: %v", err)
	}
	if err := repo.Create(&model.TimelineNode{ApplicationID: app.ID, Title: "far", DueDate: time.Now().AddDate(0,0,30)}); err != nil {
		t.Fatalf("create node: %v", err)
	}
	all, _ := repo.ListAll()
	svc := NewTimelineService(repo, slog.Default())
	items := svc.Upcoming(all, 7)
	if len(items) != 1 || items[0].Title != "due" {
		t.Fatalf("Upcoming() = %+v, want one node 'due'", items)
	}
}


func TestG2(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	node := &model.TimelineNode{ApplicationID: app.ID, Title: "a"}
	if err := repo.Create(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	svc := NewTimelineService(repo, slog.Default())
	got, err := svc.MarkDone(node.ID)
	if err != nil {
		t.Fatalf("MarkDone() error = %v", err)
	}
	if !got.IsDone {
		t.Fatalf("MarkDone() returned IsDone=false, want true")
	}
}


func TestG3(t *testing.T) {
	d := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	got := util.FormatDate(d)
	if got != "2026-08-21" {
		t.Fatalf("FormatDate() = %q, want %q", got, "2026-08-21")
	}
}


func TestG4(t *testing.T) {
	got := util.AppStatusText("waiting")
	if got != "等待结果" {
		t.Fatalf("AppStatusText(waiting) = %q, want %q", got, "等待结果")
	}
}


func TestG5(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	svc := NewTimelineService(repo, slog.Default())
	_, err := svc.Create(app.ID, &model.TimelineNode{Title: "", DueDate: time.Now().AddDate(0,0,1)})
	if err == nil {
		t.Fatalf("Create() with empty title should return error")
	}
}


func TestG6(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	svc := NewTimelineService(repo, slog.Default())
	_, err := svc.Create(app.ID, &model.TimelineNode{Title: "a"})
	if err == nil {
		t.Fatalf("Create() with zero due date should return error")
	}
}
