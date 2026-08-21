# BUG_REPRO

推荐院校 ID 切片解析错误。触发方式：运行 backend/internal/f 的 TestF1~TestF5。错误信息：列表顺序错；UniversityIDs 格式为空格分隔；解析返回 0 个院校；CounselorID 丢失；空 ID 未拒绝。
