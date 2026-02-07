# 统一文件存储服务

## 问题陈述

后端开发团队在每个新服务中都需要重复实现文件存储功能，导致大量重复劳动和代码冗余。在影视行业场景下，需要处理大量视频文件（几 GB 到几十 GB）的上传和存储，同时必须通过预签名 URL 保证文件访问的安全性，避免使用公开链接。当前缺乏统一的企业级文件存储服务，导致开发效率低下，代码质量参差不齐。

## 证据

- **观察**：项目代码库中完全没有文件存储实现，每次需要文件功能都要从零开发
- **行业特性**：影视行业需要处理大文件（视频素材），普通 HTTP 上传无法满足需求
- **安全需求**：必须使用预签名 URL 避免文件公开访问，保护版权内容
- **假设**：需要通过实际接入验证统一接口是否真正减少重复开发

## 提议的解决方案

构建一个企业级的统一文件存储服务，通过抽象的 Storage 接口层适配云存储后端（MVP 阶段支持 AWS S3），提供预签名 URL 生成、分片上传、文件元数据管理等核心能力。服务采用 Go 语言实现，遵循项目现有的分层架构（Handler → Service → Repository），后端开发可以通过统一的 API 接口快速接入文件功能，无需重复实现存储逻辑。

## 关键假设

我们相信**统一的存储接口和预签名 URL 能力**将**消除后端开发的重复劳动**对于**企业内部的后端开发团队**。我们会通过**新服务接入时不再需要实现存储逻辑，代码复用率达到 100%**来验证我们是对的。

## 我们不做什么

- **文件版本管理** - v1 不支持文件历史版本，只保留最新版本
- **文件预览/转码** - 不提供视频转码、图片压缩等媒体处理能力，由业务方自行处理
- **文件搜索** - 不提供全文搜索或标签搜索，只支持通过文件 ID 精确查询
- **访问控制（ACL）** - 不提供细粒度的权限控制，所有访问通过预签名 URL 的过期时间控制
- **断点续传** - v1 不支持，v2 再实现（先用分片上传解决大文件问题）
- **范围下载** - v1 不支持部分内容下载，v2 再实现

## 成功指标

| 指标 | 目标 | 如何衡量 |
|------|------|----------|
| 代码复用率 | 100% | 新服务接入文件功能时，存储逻辑代码行数为 0 |
| 接入时间 | < 1 小时 | 从开始集成到完成首次文件上传的时间 |
| 大文件上传成功率 | > 95% | 1GB 以上文件上传成功次数 / 总尝试次数 |
| 预签名 URL 生成延迟 | < 100ms | P99 延迟 |

## 待解决问题

- [ ] 云存储成本预算和配额管理策略（需要运维确认）
- [ ] 预签名 URL 的默认过期时间（需要安全团队评审）
- [ ] 分片上传的 part size 默认值（需要性能测试验证）

---

## 用户与场景

**主要用户**
- **谁**：企业内部的后端开发工程师
- **当前行为**：每次需要文件功能时，从零实现 S3 SDK 集成、上传逻辑、元数据存储
- **触发时机**：开发新服务需要文件上传功能，或为现有服务添加文件管理能力
- **成功状态**：通过调用统一的 API 接口，10 分钟内完成文件上传功能集成

**要完成的任务**
当我需要为服务添加文件上传功能时，我想要直接调用统一的存储服务 API，这样我就能避免重复实现存储逻辑，专注于业务功能开发。

**非用户**
- **终端用户（前端/移动端）**：不直接调用此服务，通过业务服务的 API 间接使用
- **外部合作伙伴**：不对外开放，仅限企业内部使用
- **运维人员**：不是主要用户，但需要配置和监控服务

---

## 解决方案细节

### 核心能力（MoSCoW）

| 优先级 | 能力 | 理由 |
|--------|------|------|
| Must | S3 适配层（统一接口） | 屏蔽云存储差异，为未来支持多云打基础 |
| Must | 小文件上传（直接上传 + 预签名 URL） | 处理缩略图、字幕、海报等小文件（< 100MB），无需分片 |
| Must | 大文件分片上传（Multipart Upload） | 影视行业视频文件场景必需，普通上传无法处理 GB 级文件 |
| Must | 预签名 URL 生成（下载） | 核心安全需求，避免文件公开访问 |
| Must | 文件元数据管理 | 存储文件信息（ID、名称、大小、哈希值等）到数据库 |
| Should | 本地存储适配（开发环境） | 方便开发调试，无需依赖云存储 |
| Could | 预签名 URL 缓存（Redis） | 减少重复生成，提升性能 |
| Won't | 断点续传 | v2 实现，v1 用分片上传解决大文件问题 |
| Won't | 范围下载（Range Download） | v2 实现，v1 只支持完整文件下载 |
| Won't | 多云同时支持（OSS/COS） | v2 实现，v1 只支持 S3 |

### MVP 范围

**最小可验证功能**：
1. 小文件上传（< 100MB）：
   - 后端代理上传（POST 文件内容）
   - 前端直传（获取预签名 URL）
2. 大文件分片上传（>= 100MB）：
   - 初始化分片上传
   - 生成分片预签名 URL
   - 完成分片上传
3. 通过文件 ID 获取预签名下载 URL
4. 查询文件元数据（名称、大小、上传时间）
5. 删除文件（同时删除 S3 对象和数据库记录）

**验证标准**：
- 后端开发可以在 1 小时内完成接入
- 成功上传小文件（< 100MB）：缩略图、字幕、海报
- 成功上传大文件（>= 1GB）：视频素材
- 生成的预签名 URL 可以直接下载文件
- 前端可以使用预签名 URL 直接上传到 S3

### 用户流程

**核心路径 - 小文件上传（方式 1：后端代理）**：
1. 业务服务调用 `POST /api/v1/files/upload` 直接上传文件内容
2. 存储服务将文件上传到 S3，保存元数据到数据库
3. 返回文件 ID 和元数据

**核心路径 - 小文件上传（方式 2：前端直传）**：
1. 业务服务调用 `POST /api/v1/files/upload/presigned` 获取预签名上传 URL
2. 业务服务将 URL 返回给前端
3. 前端使用预签名 URL 直接上传到 S3
4. 上传完成后，前端调用 `POST /api/v1/files/upload/confirm` 确认上传
5. 存储服务保存元数据到数据库，返回文件 ID

**核心路径 - 大文件分片上传**：
1. 业务服务调用 `POST /api/v1/files/upload/multipart/init` 初始化上传（获取 upload_id 和分片预签名 URL 列表）
2. 业务服务将文件分片，使用预签名 URL 直接上传到 S3
3. 所有分片上传完成后，调用 `POST /api/v1/files/upload/multipart/complete` 完成上传
4. 存储服务返回文件 ID 和元数据

**核心路径 - 文件下载**：
1. 业务服务调用 `GET /api/v1/files/{file_id}/download-url` 获取预签名下载 URL
2. 业务服务将 URL 返回给终端用户
3. 终端用户通过预签名 URL 直接从 S3 下载文件

---

## 技术方案

**可行性**：HIGH

**架构设计**

```
pkg/storage/              # 存储抽象层（可复用）
├── storage.go            # Storage 接口定义
├── s3/
│   ├── client.go         # S3 客户端实现
│   └── multipart.go      # 分片上传逻辑
└── local/
    └── client.go         # 本地存储实现（开发环境）

internal/
├── models/
│   └── file.go           # 文件元数据模型
├── repositories/
│   └── file_repository.go # 文件数据访问层
├── services/
│   └── file_service.go    # 文件业务逻辑
└── handlers/
    └── file_handler.go    # 文件 API 处理器
```

**关键技术决策**

1. **Storage 接口设计**
   ```go
   type Storage interface {
       // === 小文件上传（< 100MB）===

       // 直接上传（后端代理）
       Upload(ctx context.Context, key string, reader io.Reader, size int64) error

       // 生成小文件上传预签名 URL（前端直传）
       GeneratePresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error)

       // === 大文件分片上传（>= 100MB）===

       // 初始化分片上传
       InitMultipartUpload(ctx context.Context, key string) (*MultipartUpload, error)

       // 生成分片上传预签名 URL
       GeneratePresignedPartURL(ctx context.Context, uploadID string, partNumber int) (string, error)

       // 完成分片上传
       CompleteMultipartUpload(ctx context.Context, uploadID string, parts []CompletedPart) error

       // === 通用操作 ===

       // 生成下载预签名 URL
       GeneratePresignedDownloadURL(ctx context.Context, key string, expiry time.Duration) (string, error)

       // 删除对象
       Delete(ctx context.Context, key string) error
   }
   ```

2. **依赖库选择**
   - AWS SDK v2：`github.com/aws/aws-sdk-go-v2/service/s3`
   - 原因：官方维护，功能完整，支持分片上传和预签名 URL

3. **元数据存储**
   - 使用 PostgreSQL 存储文件元数据
   - 字段：id, name, size, content_type, storage_key, upload_status, created_at, updated_at

**技术风险**

| 风险 | 可能性 | 缓解措施 |
|------|--------|----------|
| 分片上传状态管理复杂 | 中 | 使用数据库事务保证一致性，记录每个分片的上传状态 |
| 预签名 URL 过期时间配置不当 | 中 | 提供可配置的过期时间，默认 15 分钟（下载）/ 1 小时（上传） |
| S3 API 调用失败（网络/限流） | 高 | 实现重试机制（指数退避），记录详细日志 |
| 大文件上传内存占用 | 低 | 使用流式处理，不将文件加载到内存 |

---

## 实施阶段

<!--
  STATUS: pending | in-progress | complete
  PARALLEL: phases that can run concurrently (e.g., "with 3" or "-")
  DEPENDS: phases that must complete first (e.g., "1, 2" or "-")
  PRP: link to generated plan file once created
-->

| # | 阶段 | 描述 | 状态 | 并行 | 依赖 | PRP 计划 |
|---|------|------|------|------|------|----------|
| 1 | 基础设施搭建 | 配置管理、数据库模型、Storage 接口定义 | complete | - | - | [阶段1完成](.claude/prp-ralph.state.md) |
| 2 | S3 适配层实现 | S3 客户端、分片上传、预签名 URL 生成 | complete | - | 1 | [阶段2完成](.claude/prp-ralph.state.md) |
| 3 | 业务逻辑层实现 | Service 层、Repository 层、元数据管理 | complete | - | 2 | [阶段3完成](.claude/prp-ralph.state.md) |
| 4 | API 层实现 | Handler 层、路由注册、Swagger 文档 | complete | - | 3 | [阶段4完成](.claude/prp-ralph.state.md) |
| 5 | 测试与文档 | 单元测试、集成测试、接入文档 | pending | - | 4 | - |

### 阶段详情

**阶段 1：基础设施搭建**
- **目标**：建立项目基础结构和配置
- **范围**：
  - 添加 S3 配置（bucket、region、credentials）
  - 定义 File 数据库模型（id, name, size, storage_key, status 等）
  - 定义 Storage 接口（InitMultipartUpload, GeneratePresignedURL 等）
  - 数据库迁移脚本
- **成功信号**：配置文件完整，数据库表创建成功，接口定义通过编译

**阶段 2：S3 适配层实现**
- **目标**：实现 S3 存储后端的核心功能
- **范围**：
  - 初始化 S3 客户端（使用 aws-sdk-go-v2）
  - 实现小文件上传：
    - Upload：直接上传（后端代理）
    - GeneratePresignedUploadURL：生成上传预签名 URL（前端直传）
  - 实现大文件分片上传：
    - InitMultipartUpload：初始化分片上传
    - GeneratePresignedPartURL：生成分片预签名 URL
    - CompleteMultipartUpload：完成分片上传
  - 实现预签名下载 URL 生成
  - 实现对象删除
  - 错误处理和重试机制
- **成功信号**：单元测试通过，可以成功上传/下载小文件和大文件到 S3

**阶段 3：业务逻辑层实现**
- **目标**：实现文件管理的业务逻辑
- **范围**：
  - FileRepository：CRUD 操作（创建、查询、更新、删除文件记录）
  - FileService：
    - UploadDirect：直接上传小文件（后端代理）
    - InitPresignedUpload：生成小文件上传预签名 URL
    - ConfirmUpload：确认前端直传完成
    - InitMultipartUpload：初始化大文件分片上传
    - CompleteMultipartUpload：完成大文件分片上传
    - GetDownloadURL：生成下载预签名 URL
    - DeleteFile：删除文件（S3 + 数据库）
  - 事务管理（确保 S3 和数据库操作一致性）
- **成功信号**：Service 层单元测试通过，业务逻辑正确

**阶段 4：API 层实现**
- **目标**：提供 RESTful API 接口
- **范围**：
  - FileHandler：
    - **小文件上传**：
      - POST /api/v1/files/upload - 直接上传（后端代理）
      - POST /api/v1/files/upload/presigned - 获取上传预签名 URL
      - POST /api/v1/files/upload/confirm - 确认前端直传完成
    - **大文件分片上传**：
      - POST /api/v1/files/upload/multipart/init - 初始化分片上传
      - POST /api/v1/files/upload/multipart/complete - 完成分片上传
    - **通用操作**：
      - GET /api/v1/files/{id}/download-url - 获取下载 URL
      - GET /api/v1/files/{id} - 获取文件元数据
      - DELETE /api/v1/files/{id} - 删除文件
  - 请求/响应结构体定义（包含 example tag）
  - Swagger 注释
  - 路由注册
- **成功信号**：API 测试通过，Swagger 文档生成正确

**阶段 5：测试与文档**
- **目标**：确保代码质量和可用性
- **范围**：
  - 单元测试（覆盖率 > 80%）
  - 集成测试（真实 S3 环境测试）
  - 性能测试（1GB 文件上传测试）
  - 接入文档（如何在业务服务中集成）
  - 配置文档（S3 配置说明）
- **成功信号**：所有测试通过，文档完整，可以交付使用

### 并行性说明

所有阶段必须顺序执行，因为存在强依赖关系：
- 阶段 2 依赖阶段 1 的接口定义和配置
- 阶段 3 依赖阶段 2 的 S3 实现
- 阶段 4 依赖阶段 3 的业务逻辑
- 阶段 5 依赖阶段 4 的完整功能

---

## 决策日志

| 决策 | 选择 | 备选方案 | 理由 |
|------|------|----------|------|
| 存储抽象层 | 自研 Storage 接口 | thanos-io/objstore, qor/oss | 现有库缺少预签名 URL 和分片上传控制，无法满足需求 |
| MVP 后端 | 仅支持 S3 | 同时支持 S3 + OSS | 降低初期复杂度，验证接口设计，v2 再扩展 |
| 分片上传 | 使用 S3 Multipart API | 自己实现分片逻辑 | S3 原生支持，稳定可靠，无需重复造轮子 |
| 元数据存储 | PostgreSQL | Redis / 对象存储元数据 | 需要持久化和事务支持，PostgreSQL 更合适 |
| SDK 选择 | aws-sdk-go-v2 | aws-sdk-go-v1, minio-go | v2 是官方推荐，性能更好，支持 context |

---

## 研究总结

**市场背景**
- 现有开源库（thanos-io/objstore, qor/oss）主要为监控系统设计，缺少企业文件服务的核心功能
- thanos-io/objstore 不支持预签名 URL 和分片上传接口暴露
- qor/oss 支持预签名 URL 但社区活跃度低，维护状态不明
- 大多数企业选择直接使用云厂商 SDK 或自研抽象层

**技术背景**
- 项目代码库是纯净的 Go Web 模板，无任何文件存储实现
- 已有完整的分层架构（Handler → Service → Repository）
- 已有数据库（GORM + PostgreSQL）和缓存（Redis）基础设施
- TODO 目录有详细的 Python 参考实现和功能设计文档

---

*生成时间：2026-02-05*
*状态：DRAFT - 待验证*
