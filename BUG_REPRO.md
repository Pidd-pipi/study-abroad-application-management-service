# BUG_REPRO

站内消息接收人与已读权限错误。触发方式：运行 backend/internal/d 的 TestD1~TestD4。错误信息：全部消息列表少一条；非接收者可已读；未读数口径错误；发送后接收人写成发送者。
