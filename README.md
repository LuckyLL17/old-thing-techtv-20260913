# 旧物改造灵感与教程平台 (Upcycle Hub)

面向手工爱好者与环保生活践行者的创意分享平台。用户可发布将废旧物品改造成新用品的教程（如旧牛仔裤改背包、玻璃瓶改台灯、木托盘改花架），其他人可浏览、收藏、评论、评分，并按教程复刻改造。理念：**让旧物重新发光**。

## 功能模块

| # | 模块 | 说明 |
|---|---|---|
| 1 | 用户模块 | 注册 / 登录 / 密码重置（JWT）；头像、昵称、擅长领域；新手/学徒/匠人/大师等级；作品集展示 |
| 2 | 改造教程 | 标题、简介、材料清单、工具清单、图文步骤、对比前后首图、难度（简单/中等/困难）、耗时估算、分类、草稿/发布/归档状态、版本历史、版本快照与回滚、浏览/收藏/尝试计数、标签 |
| 3 | 步骤编辑器 | 顺序拖拽重排、文字+图、步骤提醒、步骤耗时 |
| 4 | 作品展示 | 上传作品图、关联教程、个性化改动、受欢迎评分、评论交流 |
| 5 | 互动社区 | 教程评论/回复、作品点赞收藏、用户关注与私信、通知中心（评论/回复/收藏/关注/尝试/作品/系统/审核结果） |
| 6 | 发现推荐 | 分类浏览、难度筛选、最新/热门/最多尝试、热门标签云、随机灵感、关键词搜索 |
| 7 | 个人中心 | 我发布的 / 收藏 / 尝试记录 / 作品 / 改造总计 / 站内通知 |
| 8 | 统计看板 + 管理 | 平台总数、TOP10 热门教程、最活跃用户、分类占比、月度趋势、审计日志、管理员操作审计统计 |

## 技术栈

| 层级 | 选型 |
|---|---|
| 后端框架 | Gin（HTTP 路由） |
| ORM | GORM（SQLite 驱动，便于本地零依赖运行） |
| 认证 | JWT（HS256，可配置过期时间） |
| 日志 | Zap（结构化日志，控制台+文件可选） |
| 配置 | Viper（YAML + 环境变量覆盖） |
| 限流 | 滑动窗口内存限流器（可开关 / 配置窗口与阈值） |
| 文件处理 | 图片自动缩放 + 缩略图生成 |
| 前端 | HTML + CSS + 原生 JS 单页（SPA 前端资源目录自动回退到 index.html） |
| 构建 | Go 1.21+ 原生多阶段构建 + Dockerfile / docker-compose |

## 代码规模（已通过 `go build` 与 0 诊断）

| 指标 | 数值 |
|---|---|
| Go 源文件 | 59 个 |
| 物理总代码行 | 5,971 行 |
| 有效 Go 代码行（去空行/纯注释） | 5,410 行 |
| 前端（HTML/CSS/JS） | 651 行 |
| SQLite 迁移脚本 | 6 个 SQL |
| 数据库表 | 16 张（users / categories / tags / tutorials / tutorial_versions / tutorial_tags / steps / materials / tools / projects / comments / favorites / attempts / follows / messages / notifications / audit_logs） |

## 目录结构

```
old-thing-techtv/
├── cmd/server/main.go                         # 入口：配置加载、依赖组装、优雅停机
├── config/
│   ├── config.go                              # Viper 配置加载
│   └── config.yaml                            # 默认配置（端口/JWT/DB/上传/限流）
├── internal/
│   ├── domain/                                # 13 个领域模型 + 常量
│   ├── repository/                            # 13 个仓储（CRUD / 计数 / 事务替换）
│   ├── service/                               # 11 个业务服务（含通知/审计/版本历史）
│   ├── handler/                               # 6 个 HTTP Handler（含通知/审计/教程历史）
│   ├── middleware/                            # 4 个中间件（Auth/OptionalAuth/CORS/Logger/RateLimit）
│   └── worker/stats_updater.go                # 定时任务：10min 刷新计数/用户等级
├── pkg/
│   ├── errors/errors.go                       # 12 个业务错误码与 Wrap 语义
│   ├── logger/logger.go                       # Zap 封装
│   └── utils/                                 # 文件/图片/Markdown/字符串工具
├── api/
│   ├── router.go                              # 路由注册 + SPA 前端回退 + 上传接口
│   └── dto/                                   # User/Tutorial/Project 3 套请求/响应 DTO
├── migrations/                                # 6 个 SQL 迁移脚本
├── frontend/
│   ├── index.html                             # SPA 容器：首页/教程库/详情/作品/统计/编辑器/个人中心
│   ├── app.css                                # 样式
│   └── app.js                                 # 路由与渲染
├── go.mod / go.sum
├── Dockerfile                                 # Alpine 多阶段构建
├── docker-compose.yml                         # app + persistent volume
└── system.md                                  # 需求与规格说明
```

## 快速开始

### 环境要求

- **Go** ≥ 1.21
- **SQLite** 3.x（GORM 自带 `modernc.org/sqlite` 纯 Go 驱动，无需本机 `sqlite3` 命令）
- 可选 **Docker / Docker Compose**

### 本地运行

```bash
# 1. 克隆并进入项目
cd old-thing-techtv

# 2. 拉取依赖（国内推荐指定代理）
GOPROXY=https://proxy.golang.org,direct go mod tidy

# 3. 构建
go build -o upcycle-hub ./cmd/server

# 4. 启动（默认读取 config/config.yaml，可传参覆盖）
./upcycle-hub config/config.yaml
```

启动后访问：

- 前端页面：http://localhost:8080
- 健康检查：http://localhost:8080/api/v1/health
- 上传目录：项目根目录下 `./uploads`（可在 `config.yaml` 中修改）
- 数据库文件：项目根目录下 `./upcycle.db`（首次启动自动建表，AutoMigrate）

### Docker 方式

项目根目录已提供 [Dockerfile](file:///Users/tog_11/code/我的go/30个/old-thing-techtv/Dockerfile)（Alpine 多阶段构建，产物约 22MB）与 [docker-compose.yml](file:///Users/tog_11/code/我的go/30个/old-thing-techtv/docker-compose.yml)（两个命名卷持久化 + 健康检查 + 环境变量覆盖）。

#### 1. 一键启动（推荐）

```bash
# 0. 前置：确保 Docker Engine 已运行，且已安装 Docker Compose v2（随 Docker Desktop 自带）
docker compose version

# 1. 可选：生产环境务必自定义 JWT 密钥
export UPCYCLE_JWT_SECRET="$(openssl rand -hex 32)"
echo "UPCYCLE_JWT_SECRET=$UPCYCLE_JWT_SECRET" > .env

# 2. 构建并后台启动
docker compose up -d --build

# 3. 查看启动日志 + 等待健康检查
docker compose logs -f app
# 日志中出现 "🚀 旧物改造平台启动于 http://0.0.0.0:8080" 即表示已就绪

# 4. 检查健康状态
docker compose ps
# 期望看到 STATE 列显示 Up (healthy)

# 5. 浏览器访问
open http://localhost:8080
# 健康检查端点：http://localhost:8080/api/v1/health
```

#### 2. 只构建镜像、不在 compose 中运行

```bash
# 构建镜像
docker build -t upcycle-hub:latest .

# 直接 docker run（把数据挂到本地目录，更便于查看）
mkdir -p ./data/db ./data/uploads
docker run -d --name upcycle-hub \
  -p 8080:8080 \
  -e UPCYCLE_JWT_SECRET="生产环境请替换为长随机串" \
  -e UPCYCLE_SERVER_MODE=release \
  -e UPCYCLE_LOG_LEVEL=info \
  -v "$(pwd)/data/db:/data/db" \
  -v "$(pwd)/data/uploads:/data/uploads" \
  --restart unless-stopped \
  upcycle-hub:latest
```

#### 3. 持久化与数据位置

Compose 使用两个命名卷（**容器删除后数据依然保留**）：

| 卷名 | 容器内路径 | 内容 |
|---|---|---|
| `upcycle-db` | `/data/db` | SQLite 数据库文件 `upcycle.db` |
| `upcycle-uploads` | `/data/uploads` | 上传的图片、缩略图 |

若想改为挂载到本地目录（便于备份/迁移），把 `docker-compose.yml` 末尾 `volumes:` 段改为：

```yaml
services:
  app:
    volumes:
      - ./data/db:/data/db
      - ./data/uploads:/data/uploads
# 顶部不要声明命名卷
```

备份示例：

```bash
# 导出数据库文件（命名卷方式）
docker run --rm -v upcycle-db:/src -v "$(pwd):/dst" alpine cp /src/upcycle.db /dst/

# 全量打包（命名卷 + compose 文件）
tar czf upcycle-backup-$(date +%F).tar.gz docker-compose.yml .env \
  $(docker run --rm -v upcycle-db:/src -w /src alpine sh -c 'tar c .' > /dev/null; echo '需要用命名卷命令单独导出')

# 或者直接用官方备份命令（推荐）
docker run --rm -v upcycle-db:/data -v "$(pwd):/backup" alpine \
  tar czf /backup/upcycle-db-$(date +%F).tar.gz -C /data .
docker run --rm -v upcycle-uploads:/data -v "$(pwd):/backup" alpine \
  tar czf /backup/upcycle-uploads-$(date +%F).tar.gz -C /data .
```

#### 4. Compose 可配置环境变量（均可覆盖 config/config.yaml 默认值）

| 环境变量 | 默认值 | 说明 |
|---|---|---|
| `UPCYCLE_SERVER_HOST` | `0.0.0.0` | 监听地址，容器内无需修改 |
| `UPCYCLE_SERVER_PORT` | `8080` | 容器内端口，映射需同时修改 `ports:` |
| `UPCYCLE_SERVER_MODE` | `release` | `debug` / `release` / `test` |
| `UPCYCLE_DB_DRIVER` | `sqlite` | 容器版固定 SQLite |
| `UPCYCLE_DB_DSN` | `/data/db/upcycle.db` | 持久化库文件，请勿改到非 `/data` 外路径 |
| `UPCYCLE_JWT_SECRET` | 占位随机串 | **生产必须设置强随机值** |
| `UPCYCLE_JWT_EXPIREHOUR` | `24` | 登录令牌有效期（小时） |
| `UPCYCLE_LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `UPCYCLE_UPLOAD_DIR` | `/data/uploads` | 上传存储路径，必须与卷一致 |
| `UPCYCLE_UPLOAD_MAXSIZE` | `10485760` | 单文件字节上限（默认 10MB） |
| `UPCYCLE_UPLOAD_IMAGEMAX` | `1920` | 图片自动缩放到的最大边长 |
| `UPCYCLE_RATE_ENABLE` | `true` | 是否启用限流 |
| `UPCYCLE_RATE_LIMIT` | `300` | 每窗口最大请求数（按 IP） |
| `UPCYCLE_RATE_WINDOW` | `60` | 限流窗口，秒 |

#### 5. 常用运维命令

```bash
# 停止服务（保留数据）
docker compose stop

# 停止 + 删除容器与网络（卷保留）
docker compose down

# 彻底删除（⚠️ 包括持久化卷，数据全部丢失）
docker compose down -v

# 升级到新代码（重新构建 + 滚动重启，数据保留）
git pull
docker compose up -d --build

# 只重启 app 服务（如改了环境变量）
docker compose restart app

# 进入容器 shell 调试
docker compose exec app sh
ls /data/db && ls /data/uploads

# 查看最近 100 行日志
docker compose logs --tail=100 app

# 实时看错误日志
docker compose logs -f --tail=200 app 2>&1 | grep -iE "error|warn|panic|fatal"
```

#### 6. 构建原理

[Dockerfile](file:///Users/tog_11/code/我的go/30个/old-thing-techtv/Dockerfile) 采用典型 Go 多阶段构建：

1. **阶段 1（builder）**：`golang:1.22-alpine`，挂载 `go mod` / `go build` 缓存层，启用 `CGO_ENABLED=1` 以兼容 `modernc.org/sqlite` 纯 Go 驱动；设置 `GOPROXY=https://proxy.golang.org,direct` 避免海外构建拉包失败；最终以 `-trimpath -ldflags="-s -w"` 去除符号与调试信息，体积最小化。
2. **阶段 2（runtime）**：`alpine:3.20`，仅带 `ca-certificates` 与 `tzdata`；`config/` 与 `frontend/` 拷贝进镜像；`/data` 声明为 `VOLUME`；默认暴露 8080；通过 `ENTRYPOINT` 读取 `config/config.yaml` 并支持环境变量覆盖。

#### 7. 常见问题

- **Q: `docker compose build` 卡在 `go mod download`？**
  A: 构建机器需要能访问 `proxy.golang.org`。境内可临时换镜像：在 builder 阶段把 `GOPROXY` 改成 `https://goproxy.cn,direct`，或为 Docker daemon 配置 HTTP 代理。
- **Q: 容器启动后健康检查一直 unhealthy / 返回 404？**
  A: 先看日志 `docker compose logs app`。若出现 "config/config.yaml 不存在"，说明 COPY 失败或 docker build 上下文不对；若出现 "数据库迁移失败"，检查 `/data/db` 目录是否可写。
- **Q: 在 Apple Silicon / ARM 机器上构建后部署到 x86 服务器运行失败？**
  A: 为目标架构交叉构建：
  ```bash
  docker buildx build --platform linux/amd64 -t upcycle-hub:amd64 --load .
  # 或在 compose 中指定
  docker compose up -d --build --build-arg=BUILDPLATFORM=linux/amd64
  ```
- **Q: 想改用 PostgreSQL / MySQL？**
  A: 当前代码默认启用 SQLite。若切到其他数据库，需：①修改 `go.mod` 引入对应驱动；②替换 `cmd/server/main.go` 中的 `sqlite.Open`；③在 compose 中新增对应数据库 service 与网络；④把 `UPCYCLE_DB_DSN` 改为 `postgres://...` 形式。

## 配置文件（config/config.yaml）

```yaml
server:
  host: 0.0.0.0              # 监听地址
  port: 8080                 # 端口
  mode: debug                # gin mode: debug / release / test
db:
  driver: sqlite
  dsn: upcycle.db            # SQLite 文件路径
jwt:
  secret: "old-thing-techtv-secret-2026"  # 生产环境务必替换
  expirehour: 24             # 令牌过期小时数
log:
  level: debug               # debug/info/warn/error
  file: ""                   # 留空输出到控制台
upload:
  dir: ./uploads             # 上传保存目录
  maxsize: 10485760          # 单文件上限（字节，默认 10MB）
  imagemax: 1920             # 图片最大边长，超出自动缩放
rate:
  enable: true               # 开启限流
  limit: 300                 # 每窗口最大请求数
  window: 60                 # 窗口（秒）
```

## 核心 API

公共前缀：`/api/v1`，Body 使用 `application/json`，鉴权使用 `Authorization: Bearer <token>`。

### 认证 / 用户

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| POST | `/auth/register` | 注册 | 否 |
| POST | `/auth/login` | 登录，返回 `access_token` | 否 |
| GET  | `/auth/me` | 获取当前用户资料 | 是 |
| PUT  | `/auth/me` | 更新昵称/头像/擅长领域 | 是 |
| POST | `/auth/reset-password` | 重置密码（旧→新） | 是 |
| GET  | `/auth/center` | 个人中心改造统计 | 是 |

### 教程

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| GET  | `/tutorials` | 列表：支持分类/难度/排序/标签筛选、分页 | 否 |
| GET  | `/tutorials/:id` | 详情：步骤+材料+工具+评论计数（可选登录以标记收藏/尝试） | 可选 |
| POST | `/tutorials` | 发布教程，含 `steps/materials/tools/tags` | 是 |
| PUT  | `/tutorials/:id` | 更新（仅作者可改） | 是 |
| DELETE | `/tutorials/:id` | 删除（软删，仅作者/管理员） | 是 |
| POST | `/tutorials/:id/reorder` | 拖拽重排步骤顺序 | 是 |
| POST | `/tutorials/:id/comments` | 评论教程，支持 `parent_id` 回复 | 是 |
| GET  | `/tutorials/:id/comments` | 评论列表分页 | 否 |
| POST | `/tutorials/:id/attempt` | 标记/取消"我开始尝试这个教程" | 是 |
| POST | `/tutorials/:id/projects` | 在此教程下上传复刻作品 | 是 |
| GET  | `/tutorials/:id/history` | 教程版本历史列表 | 否 |
| GET  | `/tutorials/:id/history/:version` | 查看某一版本快照内容（JSON） | 否 |
| POST | `/tutorials/:id/history/snapshot` | 手动触发版本快照保存 | 是 |
| POST | `/tutorials/:id/history/rollback` | `{ "version": N }` 回滚到指定版本 | 是 |

### 作品

| 方法 | 路径 | 说明 | 鉴权 |
|---|---|---|---|
| GET  | `/projects` | 作品列表，可按教程/作者筛选 | 否 |
| GET  | `/projects/:id` | 作品详情与评论 | 否 |
| POST | `/projects` | 发布作品 | 是 |
| PUT/DELETE | `/projects/:id` | 改/删 | 是（作者） |
| POST | `/projects/:id/like` | 点赞 | 否 |
| POST | `/projects/:id/comments` | 评论作品 | 是 |

### 发现 / 统计

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/home` | 首页聚合：热门/最新/最多尝试/热门标签/分类/随机灵感 |
| GET | `/random` | 随机灵感：若干教程 |
| GET | `/top` | TOP 热门教程 |
| GET | `/search` | 关键词搜索（标题/摘要/标签/步骤） |
| GET | `/categories` | 分类列表（家具/服饰/容器/装饰/工具/电子/其他） |
| GET | `/tags` | 标签列表，支持 `limit` |
| GET | `/stats` | 统计看板：总数 + TOP10 + 分类占比 + 月新增趋势 |

### 个人中心 `/me`（全需鉴权）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET  | `/favorites` | 我的收藏（支持 `type=tutorial|project`） |
| POST | `/favorites` | `{ target_type, target_id }` 切换收藏 |
| GET  | `/attempts` | 我的尝试列表 |
| POST | `/follow` | `{ target_id }` 关注用户 |
| DELETE | `/follow/:id` | 取消关注 |
| POST | `/messages` | 发送私信 `{ receiver_id, content }` |
| GET  | `/messages/:id` | 与某用户的会话（分页） |
| GET  | `/unread` | 未读私信汇总 |
| GET  | `/notifications` | 站内通知列表，`only_unread=true` 只看未读 |
| GET  | `/notifications/unread-count` | 未读通知数 |
| POST | `/notifications/:id/read` | 标记单条已读 |
| POST | `/notifications/read-all` | 全部标已读，返回更新条数 |
| DELETE | `/notifications` | 清理，`older_than_days=30` 清 30 天前，留空清全部 |

### 管理端 `/admin`（鉴权 + 权限）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/audit` | 审计日志：支持 `user_id/action/target_type/from/to` 过滤分页 |
| GET | `/audit/stats` | 按动作聚合计数，`days=30` 指定统计窗口 |

### 上传

```
POST /api/v1/upload   Content-Type: multipart/form-data   Form-Field: file
```

返回：

```json
{
  "code": 0, "success": true,
  "data": {
    "path":  "/uploads/xxxx.jpg",
    "thumb": "/uploads/xxxx_thumb.jpg",
    "size":  234567
  }
}
```

## 统一响应格式

成功：

```json
{ "code": 0, "success": true, "message": "ok", "data": { ... } }
```

分页（`PageOK`）：

```json
{
  "code": 0, "success": true,
  "data": [ ... ],
  "page": 1, "size": 20, "total": 348, "pages": 18
}
```

失败：

```json
{ "code": 40101, "success": false, "message": "未登录或令牌过期" }
```

常见错误码：

| 错误码 | 含义 |
|---|---|
| 40000 | 请求参数错误 / BadRequest |
| 40101 | 未认证 / 令牌无效 |
| 40301 | 无权限 / Forbidden |
| 40401 | 资源不存在 / NotFound |
| 40901 | 冲突（例如已注册、已收藏） |
| 42200 | 业务校验失败 |
| 42901 | 请求超过限流 |
| 50002 | 数据库操作失败 |
| 50003 | 文件存储失败 |

## 单一职责与分层约定

- **domain**：纯结构 + 常量 + 表名，无业务逻辑。
- **repository**：仅做 CRUD / 计数 / 排序 / 过滤，不包含业务编排；错误统一经 `apperr.Wrap` 标注来源。
- **service**：组合多个仓储完成业务，负责权限校验、事务边界、数据一致性（如回滚时先替换步骤，再重建材料/工具，最后重拍快照）。
- **handler**：参数绑定、鉴权、调用 service、按统一结构输出；`Fail/OK/PageOK` 公共函数复用。
- **middleware**：与业务无关的横切关注点（CORS/日志/JWT/限流）。
- **worker**：`StatsUpdater` 每 10 分钟一次，对 `users.tutorial_count / users.level / categories.post_count / tags.use_count` 做最终一致性回填（避免高频写路径争用）。

## 数据库迁移说明

默认使用 `GORM AutoMigrate` 在首次启动时建表，无需手动执行 SQL。若需纯 SQL 版本，`migrations/` 下提供了 6 个顺序脚本，可按需改造 PostgreSQL / MySQL。

```
migrations/
  001_users.sql
  002_categories_tags.sql
  003_tutorials.sql
  004_steps_materials_tools.sql
  005_projects_comments_favorites_attempts.sql
  006_follows_messages_notifications_audit.sql
```

## 启动命令（快速备忘）

```bash
# 本地构建
cd old-thing-techtv
go build -o upcycle-hub ./cmd/server
./upcycle-hub config/config.yaml

# 只看日志
./upcycle-hub 2>&1 | jq -R .

# Docker 一键起
docker-compose up -d
docker-compose logs -f app
```
