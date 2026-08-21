# BUG_REPRO

申请状态机转换缺失/列表越权/创建未绑定学生。触发方式：运行 backend/internal/a 的 TestA1~TestA6。错误信息：ValidApplicationStatuses 缺 waitlisted；submitted 不能转 waiting；学生列表多返回；非 owner 可改状态；waiting 不能转 waitlisted；Create 后 StudentID 为 0。
