package service

import (
	"log/slog"
	"testing"
	"time"

	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func TestTimelineListAll(t *testing.T) {
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

func TestTimelineMarkReminderSent(t *testing.T) {
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

func TestNotificationScanMarksReminderSent(t *testing.T) {
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

func TestNotificationScanSendsToStudent(t *testing.T) {
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

func TestNotificationScanRejectsNegativeDays(t *testing.T) {
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

func TestNotificationScanRejectsNilTimelineService(t *testing.T) {
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
