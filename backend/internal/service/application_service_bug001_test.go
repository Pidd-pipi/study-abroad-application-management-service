package service

import (
	"testing"

	"github.com/gbstudyapply/gbstudyapply/internal/constants"
	"github.com/gbstudyapply/gbstudyapply/internal/model"
)

func TestApplicationValidStatusesContainWaitlisted(t *testing.T) {
	statuses := constants.ValidApplicationStatuses()
	if !containsString(statuses, constants.AppStatusWaitlisted) {
		t.Fatalf("ValidApplicationStatuses() = %v, want contain %q", statuses, constants.AppStatusWaitlisted)
	}
}

func TestApplicationSubmittedTransitionsToWaiting(t *testing.T) {
	next := constants.NextApplicationStatuses(constants.AppStatusSubmitted)
	if !containsString(next, constants.AppStatusWaiting) {
		t.Fatalf("NextApplicationStatuses(%q) = %v, want contain %q", constants.AppStatusSubmitted, next, constants.AppStatusWaiting)
	}
}

func TestApplicationServiceListStudentScope(t *testing.T) {
	db := newTestDB(t)
	svc := newApplicationService(db)
	u1 := createUser(t, db, "stu1", constants.RoleStudent)
	u2 := createUser(t, db, "stu2", constants.RoleStudent)
	uni := createUniversity(t, db)
	app1 := createApplication(t, db, u1.ID, uni.ID)
	createApplication(t, db, u2.ID, uni.ID)

	items, err := svc.List(u1.ID, constants.RoleStudent)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("List() returned %d projects, want 1", len(items))
	}
	if items[0].ID != app1.ID {
		t.Fatalf("List() first project ID = %d, want %d", items[0].ID, app1.ID)
	}
}

func TestApplicationServiceUpdateStatusOwnerCheck(t *testing.T) {
	db := newTestDB(t)
	svc := newApplicationService(db)
	u1 := createUser(t, db, "stu1", constants.RoleStudent)
	u2 := createUser(t, db, "stu2", constants.RoleStudent)
	uni := createUniversity(t, db)
	app := createApplication(t, db, u1.ID, uni.ID)

	_, err := svc.UpdateStatus(app.ID, u2.ID, constants.RoleStudent, constants.AppStatusPreparing)
	if err == nil {
		t.Fatalf("UpdateStatus() with non-owner should return error, got nil")
	}
}

func TestApplicationWaitingTransitionsToWaitlisted(t *testing.T) {
	next := constants.NextApplicationStatuses(constants.AppStatusWaiting)
	if !containsString(next, constants.AppStatusWaitlisted) {
		t.Fatalf("NextApplicationStatuses(%q) = %v, want contain %q", constants.AppStatusWaiting, next, constants.AppStatusWaitlisted)
	}
}

func TestApplicationServiceCreateStudentID(t *testing.T) {
	db := newTestDB(t)
	svc := newApplicationService(db)
	stu := createUser(t, db, "stu1", constants.RoleStudent)
	uni := createUniversity(t, db)
	created, err := svc.Create(stu.ID, &model.ApplicationProject{UniversityID: uni.ID, Major: "CS"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.StudentID != stu.ID {
		t.Fatalf("StudentID = %d, want %d", created.StudentID, stu.ID)
	}
}
