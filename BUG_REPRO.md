# BUG_REPRO

时间线到期过滤与日期格式化错误。触发方式：运行 backend/internal/g 的 TestG1~TestG6。错误信息：Upcoming 返回远未来节点；MarkDone 返回 IsDone=false；日期格式少零；waiting 文案错；空标题/零截止日期未拒绝。
