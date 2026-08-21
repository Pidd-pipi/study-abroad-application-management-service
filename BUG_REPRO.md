# BUG_REPRO

提醒扫描状态/接收人错误。触发方式：运行 backend/internal/e 的 TestE1~TestE6。错误信息：节点 ListAll 少一条；MarkReminderSent 改错字段；扫描后 ReminderSent 仍 false；学生收不到；负天数/空服务未拒绝。
