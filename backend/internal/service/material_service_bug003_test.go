package service

import (
	"log/slog"
	"testing"

	"github.com/gbstudyapply/gbstudyapply/internal/constants"
	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func TestMaterialValidStatusesContainApproved(t *testing.T) {
	if !containsString(constants.ValidMaterialStatuses(), constants.MaterialApproved) {
		t.Fatalf("ValidMaterialStatuses() = %v, want contain %q", constants.ValidMaterialStatuses(), constants.MaterialApproved)
	}
}

func TestMaterialProgressCountsOnlyCompleted(t *testing.T) {
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

func TestMaterialUpdateStatusSetsUploadedAt(t *testing.T) {
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

func TestMaterialListByApplicationOrder(t *testing.T) {
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

func TestMaterialUploadRequiresFileURL(t *testing.T) {
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
