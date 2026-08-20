package e

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



func TestE1(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	if err := repo.Create(&model.TimelineNode{ApplicationID: app.ID, Title: "a", IsDone: false}); err != nil {
		t.Fatalf("create node: %v", err)
	}
	if err := repo.Create(&model.TimelineNode{ApplicationID: app.ID, Title: "b", IsDone: true}); err != nil {
		t.Fatalf("create node: %v", err)
	}
	items, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListAll() = %d, want 2", len(items))
	}
}


func TestE2(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	node := &model.TimelineNode{ApplicationID: app.ID, Title: "a", DueDate: time.Now().AddDate(0,0,1), IsDone: false, ReminderSent: false}
	if err := repo.Create(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	if err := repo.MarkReminderSent(node.ID); err != nil {
		t.Fatalf("MarkReminderSent() error = %v", err)
	}
	got, _ := repo.FindByID(node.ID)
	if !got.ReminderSent || got.IsDone {
		t.Fatalf("MarkReminderSent() = reminder=%v is_done=%v, want reminder=true is_done=false", got.ReminderSent, got.IsDone)
	}
}


func TestE3(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	app := createApplication(t, db, stu.ID, createUniversity(t, db).ID)
	tr := repository.NewTimelineNodeRepository(db)
	mr := repository.NewMessageRepository(db)
	ts := NewTimelineService(tr, slog.Default())
	svc := NewNotificationService(tr, mr, slog.Default())
	node := &model.TimelineNode{ApplicationID: app.ID, Title: "a", DueDate: time.Now().AddDate(0,0,1), IsDone: false, ReminderSent: false}
	if err := tr.Create(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	appRepo := repository.NewApplicationProjectRepository(db)
	if _, err := svc.ScanDeadlines(ts, appRepo, 7); err != nil {
		t.Fatalf("ScanDeadlines() error = %v", err)
	}
	got, _ := tr.FindByID(node.ID)
	if !got.ReminderSent {
		t.Fatalf("ReminderSent = false, want true")
	}
}


func TestE4(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	app := createApplication(t, db, stu.ID, createUniversity(t, db).ID)
	tr := repository.NewTimelineNodeRepository(db)
	mr := repository.NewMessageRepository(db)
	ts := NewTimelineService(tr, slog.Default())
	svc := NewNotificationService(tr, mr, slog.Default())
	node := &model.TimelineNode{ApplicationID: app.ID, Title: "a", DueDate: time.Now().AddDate(0,0,1), IsDone: false, ReminderSent: false}
	if err := tr.Create(node); err != nil {
		t.Fatalf("create node: %v", err)
	}
	appRepo := repository.NewApplicationProjectRepository(db)
	if _, err := svc.ScanDeadlines(ts, appRepo, 7); err != nil {
		t.Fatalf("ScanDeadlines() error = %v", err)
	}
	msgs, _ := mr.ListByReceiver(stu.ID, false)
	if len(msgs) != 1 {
		t.Fatalf("student messages = %d, want 1", len(msgs))
	}
}


func TestE5(t *testing.T) {
	db := newTestDB(t)
	tr := repository.NewTimelineNodeRepository(db)
	mr := repository.NewMessageRepository(db)
	ts := NewTimelineService(tr, slog.Default())
	svc := NewNotificationService(tr, mr, slog.Default())
	appRepo := repository.NewApplicationProjectRepository(db)
	_, err := svc.ScanDeadlines(ts, appRepo, -1)
	if err == nil {
		t.Fatalf("ScanDeadlines() with negative days should return error")
	}
}


func TestE6(t *testing.T) {
	db := newTestDB(t)
	tr := repository.NewTimelineNodeRepository(db)
	mr := repository.NewMessageRepository(db)
	svc := NewNotificationService(tr, mr, slog.Default())
	appRepo := repository.NewApplicationProjectRepository(db)
	_, err := svc.ScanDeadlines(nil, appRepo, 7)
	if err == nil {
		t.Fatalf("ScanDeadlines() with nil TimelineService should return error")
	}
}
