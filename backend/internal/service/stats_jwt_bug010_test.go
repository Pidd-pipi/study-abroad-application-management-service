package service

import (
	"testing"
	"time"

	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/util"
)

func TestComputeAppStatsTotal(t *testing.T) {
	projects := []model.ApplicationProject{{Status: "planning"}, {Status: "submitted"}}
	s := ComputeAppStats(projects)
	if s.Total != 2 {
		t.Fatalf("Total = %d, want 2", s.Total)
	}
}

func TestComputeAppStatsAdmitted(t *testing.T) {
	projects := []model.ApplicationProject{{Status: "admitted"}}
	s := ComputeAppStats(projects)
	if s.Admitted != 1 {
		t.Fatalf("Admitted = %d, want 1", s.Admitted)
	}
}

func TestComputeAppStatsAppliedIncludesWaitlisted(t *testing.T) {
	projects := []model.ApplicationProject{{Status: "waitlisted"}}
	s := ComputeAppStats(projects)
	if s.Applied != 1 {
		t.Fatalf("Applied = %d, want 1", s.Applied)
	}
}

func TestJWTGenerateAndParse(t *testing.T) {
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

func TestJWTGenerateRejectsEmptySecret(t *testing.T) {
	_, err := util.GenerateToken(1, "stu", "student", "", time.Hour)
	if err == nil {
		t.Fatalf("GenerateToken() with empty secret should return error")
	}
}

func TestJWTGenerateRejectsNonPositiveExpire(t *testing.T) {
	_, err := util.GenerateToken(1, "stu", "student", "secret", 0)
	if err == nil {
		t.Fatalf("GenerateToken() with non-positive expire should return error")
	}
}
