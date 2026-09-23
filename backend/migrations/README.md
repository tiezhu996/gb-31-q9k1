# 迁移说明

本项目的 MongoDB 索引迁移由后端启动时自动执行（`internal/database/database.go -> ensureIndexes`），
幂等创建全部唯一索引与查询索引。业务迁移脚本（如需修改历史数据）可放置于本目录，
并通过 `internal/service` 中的迁移入口调用。

当前索引清单见根目录 `README.md` 的「数据库集合与索引」小节。
