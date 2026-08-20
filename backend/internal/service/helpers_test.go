package service

import (
	"log/slog"
	"testing"

	"gorm.io/gorm"

	"github.com/gbstudyapply/gbstudyapply/internal/constants"
	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func newApplicationService(db *gorm.DB) *ApplicationService {
	appRepo := repository.NewApplicationProjectRepository(db)
	univRepo := repository.NewUniversityRepository(db)
	return NewApplicationService(appRepo, univRepo, slog.Default())
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

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
