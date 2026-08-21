# BUG_REPRO

登录密码校验与学生列表错误。触发方式：运行 backend/internal/i 的 TestI1~TestI6。错误信息：学生列表为 0；错误密码也能登录；手机号未更新；注册角色为 admin；空用户名/空邮箱未拒绝。
