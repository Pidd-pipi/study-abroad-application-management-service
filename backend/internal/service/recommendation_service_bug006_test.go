package service

import (
	"log/slog"
	"testing"

	"gorm.io/gorm"

	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func newRecommendationService(db *gorm.DB) *RecommendationService {
	return NewRecommendationService(repository.NewRecommendationRepository(db), repository.NewUniversityRepository(db), slog.Default())
}

func TestRecommendationListOrder(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	repo := repository.NewRecommendationRepository(db)
	if err := repo.Create(&model.Recommendation{StudentID: stu.ID, CounselorID: 1, UniversityIDs: "[1]"}); err != nil {
		t.Fatalf("create rec: %v", err)
	}
	if err := repo.Create(&model.Recommendation{StudentID: stu.ID, CounselorID: 1, UniversityIDs: "[2]"}); err != nil {
		t.Fatalf("create rec: %v", err)
	}
	items, err := repo.ListByStudent(stu.ID)
	if err != nil {
		t.Fatalf("ListByStudent() error = %v", err)
	}
	if len(items) != 2 || items[0].ID < items[1].ID {
		t.Fatalf("ListByStudent() order = %+v, want latest first", items)
	}
}

func TestRecommendationCreateUniversityIDsFormat(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	svc := newRecommendationService(db)
	rec, err := svc.Create(1, stu.ID, []uint{1, 2}, "reason")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if rec.UniversityIDs != "[1,2]" {
		t.Fatalf("UniversityIDs = %q, want %q", rec.UniversityIDs, "[1,2]")
	}
}

func TestRecommendationResolveUniversities(t *testing.T) {
	db := newTestDB(t)
	u1 := &model.University{Name: "U1", Country: "US", Ranking: 1}
	u2 := &model.University{Name: "U2", Country: "US", Ranking: 2}
	if err := db.Create(u1).Error; err != nil { t.Fatalf("create univ: %v", err) }
	if err := db.Create(u2).Error; err != nil { t.Fatalf("create univ: %v", err) }
	svc := newRecommendationService(db)
	rec := &model.Recommendation{StudentID: 1, CounselorID: 1, UniversityIDs: "[1,2]"}
	items, err := svc.ResolveUniversities(rec)
	if err != nil {
		t.Fatalf("ResolveUniversities() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ResolveUniversities() = %d, want 2", len(items))
	}
}

func TestRecommendationCreateCounselorID(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	svc := newRecommendationService(db)
	rec, err := svc.Create(7, stu.ID, []uint{1}, "reason")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if rec.CounselorID != 7 {
		t.Fatalf("CounselorID = %d, want 7", rec.CounselorID)
	}
}

func TestRecommendationCreateRequiresUniversityIDs(t *testing.T) {
	db := newTestDB(t)
	stu := createUser(t, db, "stu1", "student")
	svc := newRecommendationService(db)
	_, err := svc.Create(1, stu.ID, nil, "reason")
	if err == nil {
		t.Fatalf("Create() with empty university IDs should return error")
	}
}
