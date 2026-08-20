package service

import (
	"log/slog"
	"testing"

	"gorm.io/gorm"
	"time"

	"github.com/gbstudyapply/gbstudyapply/internal/config"
	"github.com/gbstudyapply/gbstudyapply/internal/constants"
	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func newUserService(t *testing.T, db *gorm.DB) *UserService {
	t.Helper()
	return NewUserService(repository.NewUserRepository(db), slog.Default(), &config.Config{JWTSecret: "secret", JWTExpire: time.Hour})
}

func TestUserListStudents(t *testing.T) {
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

func TestUserLoginWrongPassword(t *testing.T) {
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

func TestUserUpdateProfilePhone(t *testing.T) {
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

func TestUserRegisterRoleStudent(t *testing.T) {
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

func TestUserRegisterRejectsEmptyUsername(t *testing.T) {
	db := newTestDB(t)
	svc := newUserService(t, db)
	_, _, err := svc.Register("", "stu@test.com", "secret", "", "")
	if err == nil {
		t.Fatalf("Register() with empty username should return error")
	}
}

func TestUserRegisterRejectsEmptyEmail(t *testing.T) {
	db := newTestDB(t)
	svc := newUserService(t, db)
	_, _, err := svc.Register("stu1", "", "secret", "", "")
	if err == nil {
		t.Fatalf("Register() with empty email should return error")
	}
}
