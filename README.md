# 🐶 萌宠社区（gb-31 宠物图文短视频社交社区）

专属铲屎官的图文+短视频社交社区，让宠物主交流晒娃、找同城遛狗搭子。

## 项目主要功能

1. **宠物主页**：为宠物创建专属档案（品种、生日、性格、相册），宠物有独立主页与粉丝。
2. **图文/短视频动态**：发布带图/带视频动态，支持话题（#柴犬日常）、宠物标签、地点。
3. **关注与推荐**：基于品种、地点、互动行为的 Feed 推荐；关注页/发现页/同城页。
4. **实时聊天**：私信（gorilla/websocket 实时通讯），支持图片、表情包、宠物贴纸。
5. **同城遛狗搭子**：发布约伴帖，附近用户报名加入。
6. **点赞/评论/收藏/转发**：完整社交互动闭环。
7. **话题广场**：官方话题运营，参与活动榜。
8. **内容审核**：敏感词过滤 + 图片违规识别（状态机 pending/approved/rejected）。

## 快速启动（Docker Compose，首选）

```bash
cd /Users/gaobo/repositories/gitlab/评审项目/0-1代码生成提示词/golang-改编提示词/宠物主题项目提示词/gb-31
docker compose up -d --build
```

等待全部容器 healthy 后访问：

- 前端：http://localhost:8103
- 后端 API：http://localhost:3103/api/v1
- 健康检查：http://localhost:3103/healthz
- MinIO 控制台：http://localhost:47026（账号 `petsocial_minio` / `petsocial_minio_secret`）

> 演示账号：`demo / demo123`（普通用户）；`admin / admin123`（管理员）。

停止并清理：

```bash
docker compose down -v --remove-orphans
```

## 技术栈

| 层 | 技术栈 |
| --- | --- |
| 前端 | Next.js 14（App Router）+ TypeScript + Tailwind CSS + React Query + Zustand |
| 后端 | Go 1.22 + Gin + mongo-driver |
| 数据库 | MongoDB |
| 缓存 | Redis |
| 实时通信 | gorilla/websocket |
| 认证 | JWT + RBAC |
| 日志 | `log/slog` |
| 参数校验 | `github.com/go-playground/validator/v10` |
| 接口文档 | README API 清单 + `backend/api/openapi.md` |

## 项目目录结构

```
.
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go          # 环境变量集中解析
│   │   ├── database/database.go      # MongoDB 连接 + 索引
│   │   ├── database/seed.go          # 演示数据
│   │   ├── model/                    # user/pet/post/meetup/chat/audit_log 每个实体一个文件
│   │   ├── dto/                      # 每个实体一个 DTO 文件
│   │   ├── repository/               # 每个实体一个 repository 文件
│   │   ├── service/                  # 每个实体一个 service 文件 + feed/media
│   │   ├── handler/                  # 每个实体一个 handler 文件
│   │   ├── router/                   # 每个实体一个路由注册文件
│   │   ├── middleware/               # auth/rbac/request_id/error_handler/recovery/ratelimit/audit/cors/logger
│   │   ├── constants/                # error_codes/enums/messages/log_templates/errors
│   │   └── util/                     # logger/jwt/app_error/formatters/password/response/validator
│   ├── migrations/                   # 索引迁移说明
│   ├── api/openapi.md
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   └── ...
├── frontend/
│   ├── src/api/                      # 每个实体一个 API 文件
│   ├── src/components/               # StatusBadge/EmptyState/DataTable/ConfirmDialog/PostCard/PetCard/MeetupCard/Navbar/Providers
│   ├── src/app/                      # 页面（发现/关注/同城/宠物/动态/约伴/话题/私信/审计/登录/注册）
│   ├── src/stores/                   # authStore/petStore/postStore/meetupStore
│   ├── src/hooks/                    # useAuth/usePagination/useWebSocket
│   ├── src/utils/                    # request/format/auth
│   ├── src/constants/enums.ts        # 与后端对应枚举
│   ├── src/types/index.ts
│   ├── Dockerfile
│   ├── nginx.conf
│   └── package.json
├── database/                         # MongoDB 初始化脚本与说明
├── docker-compose.yml
├── .env / .env.example
└── README.md
```

## 环境变量说明

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | petsocial | Docker Compose 项目名 |
| FRONTEND_PORT | 8103 | 前端对外端口 |
| BACKEND_PORT | 3103 | 后端对外端口 |
| DB_PORT | 57604 | MongoDB 对外端口 |
| REDIS_PORT | 46321 | Redis 对外端口 |
| MINIO_PORT / MINIO_CONSOLE_PORT | 47023 / 47026 | MinIO API / 控制台端口 |
| DB_NAME / DB_USER / DB_PASSWORD | petsocial_db / petsocial_user / petsocial_pwd | MongoDB 初始化 |
| JWT_SECRET | change_me_to_a_long_random_string | JWT 签名密钥（生产必改） |
| REDIS_PASSWORD | 空 | Redis 密码 |
| MINIO_ACCESS_KEY / MINIO_SECRET_KEY | petsocial_minio / petsocial_minio_secret | MinIO 凭据 |
| SENSITIVE_WORDS | 违禁词,赌博,暴力,色情,诈骗 | 内容审核敏感词 |

## API 调用示例（curl）

注册：

```bash
curl -sS -X POST http://localhost:3103/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"xiaomei","password":"123456","nickname":"小美","city":"上海"}'
```

登录（获取 token）：

```bash
curl -sS -X POST http://localhost:3103/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo","password":"demo123"}'
```

携带 JWT 发布动态：

```bash
curl -sS -X POST http://localhost:3103/api/v1/posts \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"content":"今天带柴柴去公园#柴犬日常","type":"image","media":[{"type":"image","url":"https://example.com/a.jpg"}],"topics":["柴犬日常"],"city":"上海","location":"世纪公园"}'
```

点赞：

```bash
curl -sS -X POST http://localhost:3103/api/v1/posts/<post_id>/interact \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"type":"like"}'
```

关注用户：

```bash
curl -sS -X POST http://localhost:3103/api/v1/users/me/follow \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <token>' \
  -d '{"followee_id":"<user_id>"}'
```

## Docker 部署说明

- 端口映射：前端 `${FRONTEND_PORT:-8103}:80`，后端 `${BACKEND_PORT:-3103}:8080`，MongoDB `${DB_PORT:-57604}:27017`，Redis `${REDIS_PORT:-46321}:6379`，MinIO `${MINIO_PORT:-47023}:9000`。
- 数据卷：`mongo_data`、`redis_data`、`minio_data` 命名卷持久化，不受中文目录名影响。
- 健康检查：MongoDB/Redis/MinIO/后端均配置 healthcheck，后端 `depends_on` 等待数据库就绪。
- 常见问题：
  - 端口冲突：修改 `.env` 中对应端口后 `docker compose up -d`。
  - 首次构建较慢：npm 使用 `https://registry.npmmirror.com`，Go 使用 `https://goproxy.cn,direct` 镜像加速。
  - 重新初始化数据：`docker compose down -v` 后重新 `up -d`。

## 本地开发（备选）

后端：

```bash
cd backend
go mod tidy
go run ./cmd/server
# 构建：go build ./...
```

前端：

```bash
cd frontend
npm install
npm run dev   # 开发模式 http://localhost:3000
```

## 枚举出现位置清单

### 1. 角色 Role（user / admin）

- 后端：`internal/constants/enums.go`（定义）、`internal/model/user.go`（模型字段）、`internal/dto/user_dto.go`（DTO）、`internal/service/user_service.go`（注册默认 user）、`internal/middleware/rbac.go`（权限校验）、`internal/middleware/auth.go`（context 注入）、`internal/util/jwt.go`（token 载荷）、`internal/util/formatters.go`（FormatRole）、`internal/router/router.go`（admin 路由）、`internal/database/seed.go`（admin/demo 种子）。
- 前端：`src/constants/enums.ts`、`src/types/index.ts`、`src/utils/auth.ts`（isAdmin）、`src/components/Navbar.tsx`（审计入口显隐）、`src/hooks/useAuth.ts`、`src/app/admin/audit/page.tsx`（路由守卫）。

### 2. 动态状态 PostStatus（pending / approved / rejected）

- 后端：`internal/constants/enums.go`（定义）、`internal/model/post.go`（模型）、`internal/dto/post_dto.go`（DTO 校验 oneof）、`internal/service/post_service.go`（状态机：发布置 approved、审核流转）、`internal/handler/post_handler.go`（审核接口）、`internal/repository/post_repository.go`（List 默认过滤 approved）、`internal/util/formatters.go`（FormatStatusText）、`internal/constants/error_codes.go`（CodeContentRejected）、`internal/constants/log_templates.go`（LogPostCreated/LogPostRejected/LogPostApproved）、`internal/database/seed.go`。
- 前端：`src/constants/enums.ts`、`src/types/index.ts`、`src/components/StatusBadge.tsx`（状态徽标）、`src/components/PostCard.tsx`。

### 3. 约伴状态 MeetupStatus（open / full / cancelled / completed）

- 后端：`internal/constants/enums.go`、`internal/model/meetup.go`、`internal/dto/meetup_dto.go`（oneof 校验）、`internal/service/meetup_service.go`（状态机：open→full→cancelled/completed）、`internal/handler/meetup_handler.go`、`internal/repository/meetup_repository.go`、`internal/util/formatters.go`、`internal/constants/log_templates.go`（LogMeetupStatus 等）。
- 前端：`src/constants/enums.ts`、`src/types/index.ts`、`src/components/StatusBadge.tsx`、`src/components/MeetupCard.tsx`（按钮显隐与文案）。

### 4. 互动类型 InteractionType（like / favorite / forward）

- 后端：`internal/constants/enums.go`、`internal/model/post.go`（Interaction）、`internal/dto/post_dto.go`（oneof 校验）、`internal/service/post_service.go`（Interact 状态机与计数）、`internal/repository/post_repository.go`（唯一索引 (post_id,user_id,type)）、`internal/util/formatters.go`（FormatInteractionType）、`internal/constants/log_templates.go`（LogPostLiked/Favorited/Forwarded）。
- 前端：`src/constants/enums.ts`、`src/types/index.ts`、`src/components/PostCard.tsx`（点赞/收藏/转发按钮）。

### 5. 宠物物种 PetSpecies（dog / cat / other）与宠物性别 PetGender（male / female）

- 后端：`internal/constants/enums.go`、`internal/model/pet.go`、`internal/dto/pet_dto.go`（oneof 校验）、`internal/service/pet_service.go`、`internal/util/formatters.go`（FormatSpecies）、`internal/database/seed.go`。
- 前端：`src/constants/enums.ts`、`src/types/index.ts`、`src/components/StatusBadge.tsx`、`src/app/pets/page.tsx`（表单下拉）、`src/utils/format.ts`（formatGender）。

### 6. 私信消息类型 ChatMessageType（text / image / sticker）

- 后端：`internal/constants/enums.go`、`internal/model/chat.go`、`internal/dto/chat_dto.go`（oneof 校验）、`internal/service/chat_service.go`、`internal/constants/log_templates.go`（LogChatSent）。
- 前端：`src/constants/enums.ts`、`src/types/index.ts`、`src/app/chat/page.tsx`（图片渲染）。

## 文件结构强制清单

已按提示词要求落地（见上文目录结构）：

- 后端 `cmd/ + internal/ + pkg/` 标准布局，`handler → service → repository → model` 单向依赖、构造器注入。
- 每个实体拆分为 model / dto / repository / service / handler / router / constants 独立文件。
- 前端按模块拆分页面、组件、store、api、hooks、utils、constants。
- **严禁合并职责到单一文件**：不允许把多个实体的 model/repository/service/handler 写进同一个文件，也不允许前端把所有页面写在 App.tsx 中。

## 屎山代码设计要求（低内聚、高耦合、牵一发动全身）

1. **日志模块单独管理但全栈引用**：`internal/util/logger.go` 封装 slog，所有 handler/service/middleware 引用；日志格式集中在 `internal/constants/log_templates.go`（30+ 条模板），业务字段变更必须同步修改模板与调用处。
2. **异常信息分散且层层透传**：错误码集中在 `internal/constants/error_codes.go`，service/handler 手动拼接 message 并再次包装；错误 message 包含实体名、字段名、角色名。
3. **常量/工具类多处耦合**：`internal/util/formatters.go` 同时包含日期、状态、类型文本格式化；`internal/constants/messages.go` 同时承载接口返回文案、日志文案、错误提示文案。
4. **状态机跨多处定义**：PostStatus/MeetupStatus 的状态流转规则同时存在于 service 状态机、前端按钮显隐、日志模板、错误码、formatters。
5. **枚举多处重复定义且被多处引用**：Role/PostStatus/MeetupStatus/InteractionType/PetSpecies/ChatMessageType 同时出现在后端 constants、DTO、模型、日志模板、错误码、formatters 与前端 constants。
6. **牵一发动全身验证标准**：给核心实体新增一个字段/状态，需同步触达模型、DTO、constants、service、repository、handler、formatters、日志模板、错误码、前端类型与页面等 ≥ 10 个文件。

## License

MIT License
