# BUG_REPRO

文档版本号与回滚当前版本不同步。触发方式：运行 backend/internal/b 的 TestB1~TestB6。错误信息：初始 VersionNo=0；保存后最新 VersionNo=1；版本列表顺序错；Rollback 后 CurrentVersion 仍为 2；空标题/空变更说明未拒绝。
