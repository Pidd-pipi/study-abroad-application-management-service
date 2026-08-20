package service

import (
	"log/slog"
	"testing"

	"gorm.io/gorm"

	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func newUniversityService(db *gorm.DB) *UniversityService {
	return NewUniversityService(repository.NewUniversityRepository(db), slog.Default())
}

func TestUniversityListFiltersCountry(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	if _, err := svc.Create(&model.University{Name: "US Uni", Country: "US", Ranking: 1}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	if _, err := svc.Create(&model.University{Name: "UK Uni", Country: "UK", Ranking: 2}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	items, total, err := svc.List("US", 0, 0, "", 1, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Country != "US" {
		t.Fatalf("List(country=US) = total=%d items=%+v, want exactly one US university", total, items)
	}
}

func TestUniversityListFiltersRanking(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	if _, err := svc.Create(&model.University{Name: "Low", Country: "US", Ranking: 1}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	if _, err := svc.Create(&model.University{Name: "High", Country: "US", Ranking: 10}); err != nil {
		t.Fatalf("create univ: %v", err)
	}
	items, total, err := svc.List("", 5, 20, "", 1, 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if total != 1 || len(items) != 1 || items[0].Ranking != 10 {
		t.Fatalf("List(ranking 5-20) = total=%d items=%+v, want one university with ranking 10", total, items)
	}
}

func TestUniversityCreateDefaultsJSON(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	u, err := svc.Create(&model.University{Name: "NoJSON", Country: "US", Ranking: 1})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if u.TopMajors != "[]" || u.Requirements != "{}" {
		t.Fatalf("TopMajors=%q Requirements=%q, want [] and {}", u.TopMajors, u.Requirements)
	}
}

func TestUniversityUpdateFields(t *testing.T) {
	db := newTestDB(t)
	svc := newUniversityService(db)
	created, _ := svc.Create(&model.University{Name: "Old", Country: "US", Ranking: 1})
	updated, err := svc.Update(created.ID, &model.University{Name: "New", TuitionRange: "10-20", Requirements: "{\"gpa\":3.0}"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.TuitionRange != "10-20" {
		t.Fatalf("TuitionRange = %q, want %q", updated.TuitionRange, "10-20")
	}
}
