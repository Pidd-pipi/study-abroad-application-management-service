package service

import (
	"log/slog"
	"testing"
	"time"

	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
	"github.com/gbstudyapply/gbstudyapply/internal/util"
)

func TestTimelineUpcomingReturnsDueNodes(t *testing.T) {
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

func TestTimelineMarkDoneReturnsDone(t *testing.T) {
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

func TestFormatDate(t *testing.T) {
	d := time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC)
	got := util.FormatDate(d)
	if got != "2026-08-21" {
		t.Fatalf("FormatDate() = %q, want %q", got, "2026-08-21")
	}
}

func TestAppStatusTextWaiting(t *testing.T) {
	got := util.AppStatusText("waiting")
	if got != "等待结果" {
		t.Fatalf("AppStatusText(waiting) = %q, want %q", got, "等待结果")
	}
}

func TestTimelineCreateRejectsEmptyTitle(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	svc := NewTimelineService(repo, slog.Default())
	_, err := svc.Create(app.ID, &model.TimelineNode{Title: "", DueDate: time.Now().AddDate(0,0,1)})
	if err == nil {
		t.Fatalf("Create() with empty title should return error")
	}
}

func TestTimelineCreateRejectsZeroDueDate(t *testing.T) {
	db := newTestDB(t)
	app := createApplication(t, db, createUser(t, db, "stu1", "student").ID, createUniversity(t, db).ID)
	repo := repository.NewTimelineNodeRepository(db)
	svc := NewTimelineService(repo, slog.Default())
	_, err := svc.Create(app.ID, &model.TimelineNode{Title: "a"})
	if err == nil {
		t.Fatalf("Create() with zero due date should return error")
	}
}
