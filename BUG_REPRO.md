# BUG_REPRO

统计聚合与 JWT 生成错误。触发方式：运行 backend/internal/j 的 TestJ1~TestJ6。错误信息：总数少一；Admitted 为 0；waitlisted 不计入 Applied；JWT 刚生成即过期；空密钥/非正过期时间未拒绝。
