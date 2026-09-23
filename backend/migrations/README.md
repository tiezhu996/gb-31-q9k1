# 迁移说明

本项目的 MongoDB 索引迁移由后端启动时自动执行（`internal/database/database.go -> ensureIndexes`），
幂等创建全部唯一索引与查询索引。业务迁移脚本（如需修改历史数据）可放置于本目录，
并通过 `internal/service` 中的迁移入口调用。

当前索引清单见根目录 `README.md` 的「数据库集合与索引」小节。

## 约伴时长与报名撞车（当前版本）

- `meetups` 集合新增 `duration_minutes`（活动时长，分钟）字段；历史文档缺省按 120 分钟处理（见 `dto.DurationOrDefault` 与 service 冲突查询），无需强制回填。
- `meetups` 新增复合索引 `participants.user_id + status + meet_time`，支撑「本人进行中约伴」冲突查询。
- 新增 `meetup_join_locks` 集合：报名时按用户串行化（`_id = user_id` 唯一），`expires_at` 带 TTL 索引（0 秒过期），持有者异常退出后由后续请求抢占回收，无需人工清理。

