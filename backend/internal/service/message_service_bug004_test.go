package service

import (
	"log/slog"
	"testing"

	"github.com/gbstudyapply/gbstudyapply/internal/constants"
	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

func TestMessageListByReceiverAll(t *testing.T) {
	db := newTestDB(t)
	u := createUser(t, db, "stu1", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	if err := repo.Create(&model.Message{SenderID: 0, ReceiverID: u.ID, Content: "a", IsRead: false}); err != nil {
		t.Fatalf("create message: %v", err)
	}
	if err := repo.Create(&model.Message{SenderID: 0, ReceiverID: u.ID, Content: "b", IsRead: true}); err != nil {
		t.Fatalf("create message: %v", err)
	}
	items, err := repo.ListByReceiver(u.ID, false)
	if err != nil {
		t.Fatalf("ListByReceiver() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListByReceiver(all) = %d, want 2", len(items))
	}
}

func TestMessageMarkReadOwnerCheck(t *testing.T) {
	db := newTestDB(t)
	u1 := createUser(t, db, "stu1", constants.RoleStudent)
	u2 := createUser(t, db, "stu2", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo, slog.Default())
	msg := &model.Message{SenderID: u1.ID, ReceiverID: u2.ID, Content: "hi"}
	if err := repo.Create(msg); err != nil {
		t.Fatalf("create message: %v", err)
	}
	_, err := svc.MarkRead(u1.ID, msg.ID)
	if err == nil {
		t.Fatalf("MarkRead() by non-receiver should return error")
	}
}

func TestMessageCountUnread(t *testing.T) {
	db := newTestDB(t)
	u := createUser(t, db, "stu1", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	for _, read := range []bool{false, false, true} {
		if err := repo.Create(&model.Message{SenderID: 0, ReceiverID: u.ID, Content: "x", IsRead: read}); err != nil {
			t.Fatalf("create message: %v", err)
		}
	}
	n, err := repo.CountUnread(u.ID)
	if err != nil {
		t.Fatalf("CountUnread() error = %v", err)
	}
	if n != 2 {
		t.Fatalf("CountUnread() = %d, want 2", n)
	}
}

func TestMessageSendReceiver(t *testing.T) {
	db := newTestDB(t)
	u1 := createUser(t, db, "stu1", constants.RoleStudent)
	u2 := createUser(t, db, "stu2", constants.RoleStudent)
	repo := repository.NewMessageRepository(db)
	svc := NewMessageService(repo, slog.Default())
	msg, err := svc.Send(u1.ID, u2.ID, "hello")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if msg.ReceiverID != u2.ID {
		t.Fatalf("ReceiverID = %d, want %d", msg.ReceiverID, u2.ID)
	}
}
