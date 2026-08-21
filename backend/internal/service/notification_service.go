package service

import (
	"fmt"
	"log/slog"

	"github.com/gbstudyapply/gbstudyapply/internal/constants"
	"github.com/gbstudyapply/gbstudyapply/internal/model"
	"github.com/gbstudyapply/gbstudyapply/internal/repository"
)

// NotificationService sends deadline reminders as internal messages.
type NotificationService struct {
	timelineRepo *repository.TimelineNodeRepository
	messageRepo  *repository.MessageRepository
	logger       *slog.Logger
}

// NewNotificationService creates a NotificationService.
func NewNotificationService(timelineRepo *repository.TimelineNodeRepository, messageRepo *repository.MessageRepository, logger *slog.Logger) *NotificationService {
	return &NotificationService{timelineRepo: timelineRepo, messageRepo: messageRepo, logger: logger}
}

// ScanDeadlines finds nodes due soon, sends reminders and marks them sent.
func (s *NotificationService) ScanDeadlines(ts *TimelineService, appRepo *repository.ApplicationProjectRepository, days int) (int, error) {
	if ts == nil {
		return 0, fmt.Errorf("notification scan: nil timeline service")
	}
	if days < 0 {
		return 0, fmt.Errorf("notification scan: days must be non-negative, got %d", days)
	}
	all, err := s.timelineRepo.ListAll()
	if err != nil {
		return 0, fmt.Errorf("notification scan list: %w", err)
	}
	upcoming := ts.Upcoming(all, days)
	sent := 0
	for _, n := range upcoming {
		app, err := appRepo.FindByID(n.ApplicationID)
		if err != nil {
			s.logger.Warn("notification scan: application not found", "application_id", n.ApplicationID, "node_id", n.ID, "err", err)
			continue
		}
		if app.StudentID == 0 {
			continue
		}
		content := fmt.Sprintf("系统提醒：申请项目 #%d 的节点「%s」即将在 %s 截止，请及时准备。",
			n.ApplicationID, n.Title, n.DueDate.Format("2006-01-02"))
		msg := &model.Message{SenderID: 0, ReceiverID: app.StudentID, Content: content}
		if err := s.messageRepo.Create(msg); err != nil {
			s.logger.Error("notification scan: send message failed", "node_id", n.ID, "receiver_id", app.StudentID, "err", err)
			continue
		}
		if err := s.timelineRepo.MarkReminderSent(n.ID); err != nil {
			s.logger.Error("notification scan: mark reminder sent failed", "node_id", n.ID, "err", err)
			continue
		}
		s.logger.Info(fmt.Sprintf(constants.LogDeadlineReminderSent, n.ID), "application_id", n.ApplicationID, "receiver_id", app.StudentID)
		sent++
	}
	return sent, nil
}
