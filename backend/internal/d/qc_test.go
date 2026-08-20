package d

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



func TestD1(t *testing.T) {
	db := newTestDB(t)
	u := createUser(t, db, "stu1", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	if err := repo.Create(&model.Message{SenderID: 0, ReceiverID: u.ID, Content: "a", IsRead: false}); err != nil {
		t.Fatalf("create message: %v", err)
	}
	if err := repo.Create(&model.Message{SenderID: 0, ReceiverID: u.ID, Content: "b", IsRead: true}); err != nil {
		t.Fatalf("create message: %v", err)
	}
	items, err := repo.ListByReceiver(u.ID, false)
	if err != nil {
		t.Fatalf("ListByReceiver() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListByReceiver(all) = %d, want 2", len(items))
	}
}


func TestD2(t *testing.T) {
	db := newTestDB(t)
	u1 := createUser(t, db, "stu1", constants.RoleStudent)
	u2 := createUser(t, db, "stu2", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo, slog.Default())
	msg := &model.Message{SenderID: u1.ID, ReceiverID: u2.ID, Content: "hi"}
	if err := repo.Create(msg); err != nil {
		t.Fatalf("create message: %v", err)
	}
	_, err := svc.MarkRead(u1.ID, msg.ID)
	if err == nil {
		t.Fatalf("MarkRead() by non-receiver should return error")
	}
}


func TestD3(t *testing.T) {
	db := newTestDB(t)
	u := createUser(t, db, "stu1", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	for _, read := range []bool{false, false, true} {
		if err := repo.Create(&model.Message{SenderID: 0, ReceiverID: u.ID, Content: "x", IsRead: read}); err != nil {
			t.Fatalf("create message: %v", err)
		}
	}
	n, err := repo.CountUnread(u.ID)
	if err != nil {
		t.Fatalf("CountUnread() error = %v", err)
	}
	if n != 2 {
		t.Fatalf("CountUnread() = %d, want 2", n)
	}
}


func TestD4(t *testing.T) {
	db := newTestDB(t)
	u1 := createUser(t, db, "stu1", constants.RoleStudent)
	u2 := createUser(t, db, "stu2", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo, slog.Default())
	msg, err := svc.Send(u1.ID, u2.ID, "hello")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if msg.ReceiverID != u2.ID {
		t.Fatalf("ReceiverID = %d, want %d", msg.ReceiverID, u2.ID)
	}
}
