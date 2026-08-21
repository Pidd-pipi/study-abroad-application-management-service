# BUG_REPRO

材料状态校验/进度/上传时间/顺序错误。触发方式：运行 backend/internal/c 的 TestC1~TestC5。错误信息：approved 被判定非法；进度把 pending 计为完成；上传后 UploadedAt 为 nil；材料列表倒序；上传空地址未拒绝。
