# 数据库脚本（MongoDB）

项目使用 MongoDB 作为主数据库，集合与索引由后端启动时通过 `internal/database/database.go` 的 `ensureIndexes` 自动创建，演示数据由 `internal/database/seed.go` 自动填充。

## 集合清单

| 集合 | 说明 | 关键索引 |
| --- | --- | --- |
| users | 用户（角色字段 role: user/admin） | username 唯一 |
| follows | 用户关注关系 | (follower_id, followee_id) 唯一 |
| pets | 宠物档案 | owner_id / species / city |
| pet_follows | 宠物粉丝关系 | (pet_id, user_id) 唯一 |
| posts | 图文/短视频动态 | created_at / topics / city / status |
| comments | 评论 | (post_id, created_at) |
| interactions | 点赞/收藏/转发 | (post_id, user_id, type) 唯一 |
| topics | 话题 | name 唯一 |
| meetups | 同城遛狗搭子约伴帖 | (city, status) / participants.user_id |
| chat_messages | 私信消息 | (from_id, to_id, created_at) |
| audit_logs | 操作审计日志 | created_at / username / action |

## 初始化脚本示例

以下脚本等价于环境变量初始化方式（compose 中通过 `MONGO_INITDB_ROOT_USERNAME/PASSWORD` 完成 root 用户创建）：

```js
// database/mongo-init.js
db.getSiblingDB('petsocial_db').createUser({
  user: 'petsocial_user',
  pwd: 'petsocial_pwd',
  roles: [{ role: 'readWrite', db: 'petsocial_db' }],
});
```
