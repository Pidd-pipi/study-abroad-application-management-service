package service

import (
	"log/slog"
	"testing"

	"gorm.io/gorm"

	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func newDocumentRepos(db *gorm.DB) (*repository.DocumentRepository, *repository.DocumentVersionRepository) {
	return repository.NewDocumentRepository(db), repository.NewDocumentVersionRepository(db)
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

func TestDocumentCreateInitialVersion(t *testing.T) {
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

func TestDocumentSaveVersionNumber(t *testing.T) {
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

func TestDocumentVersionListOrder(t *testing.T) {
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

func TestDocumentRollbackUpdatesCurrentVersion(t *testing.T) {
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

func TestDocumentCreateRejectsEmptyTitle(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	docRepo, verRepo := newDocumentRepos(db)
	svc := NewDocumentService(db, docRepo, verRepo, repository.NewAnnotationRepository(db), repository.NewApplicationProjectRepository(db), slog.Default())
	_, err := svc.Create(app.ID, "ps", "", "content")
	if err == nil {
		t.Fatalf("Create() with empty title should return error")
	}
}

func TestDocumentSaveRejectsEmptySummary(t *testing.T) {
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
