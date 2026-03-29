# Prompt 1：架构设计文档

## 角色
你是一位资深的软件架构师，精通 DDD（领域驱动设计）、Go 语言最佳实践和技术文档写作。

## 目标
请阅读 `go-gateway/` 目录下的所有 Go 源代码，生成一份完善且专业的中文架构设计文档（Markdown 格式）。

## 阅读范围
请按以下顺序阅读代码，确保全面理解项目：
1. `go.mod` — 了解模块名和依赖
2. `cmd/server/main.go` — 了解入口和依赖注入
3. `configs/config.yaml` — 了解配置结构
4. `internal/infrastructure/config/config.go` — 配置加载
5. `internal/domain/` 下所有子目录（project、apikey、apiinstance、metrics）的 entity、repository、service、strategy、command
6. `internal/application/` 下的 service、dto、assembler
7. `internal/interfaces/` 下的 handler、middleware、request、response、router
8. `internal/infrastructure/persistence/` — 仓储实现
9. `internal/infrastructure/exception/errors.go` — 异常体系
10. `Dockerfile` 和 `docker-compose.yml` — 部署方案

## 文档结构要求
请严格按照以下结构输出文档：

### 1. 文档概述
- 项目名称、版本、文档修订记录模板
- 项目背景与目标（从代码中推断业务场景）

### 2. 系统架构总览
- 架构风格说明（DDD 分层架构）
- **用 Mermaid 绘制系统架构图**（展示分层关系：接口层 → 应用层 → 领域层 → 基础设施层）
- 技术栈清单（Go 版本、Gin、GORM、PostgreSQL、zerolog、go-cache 等，从 go.mod 中提取）

### 3. 目录结构说明
- 用树形结构展示项目目录
- 对每个目录/包的职责做简要说明

### 4. 领域模型设计
- 识别并描述所有限界上下文（Bounded Context）
- **用 Mermaid 绘制领域模型关系图**（实体、值对象、聚合根之间的关系）
- 对每个领域分别说明：
  - 核心实体及其字段含义
  - 枚举/值对象
  - 仓储接口定义
  - 领域服务职责

### 5. 核心业务流程
- **用 Mermaid 绘制以下流程的时序图：**
  - API 实例选择流程（从请求到返回最佳实例）
  - 调用结果上报与指标更新流程
  - API Key 认证流程
- 详细描述负载均衡策略模式的设计（策略接口 → 4种实现 → 工厂 → 亲和性装饰器）

### 6. 分层架构详解
- **接口层**：路由设计、中间件链、请求/响应模型
- **应用层**：应用服务编排逻辑、DTO 转换、降级链机制
- **领域层**：领域服务、策略模式、亲和性绑定机制
- **基础设施层**：GORM 仓储实现、配置管理、异常体系

### 7. 数据模型
- 列出所有数据库表（从 GORM 模型的 TableName() 推断）
- 描述表结构和字段含义
- **用 Mermaid 绘制 ER 图**

### 8. 部署架构
- Docker 容器化方案
- 环境变量和配置说明
- 健康检查机制

### 9. 设计决策记录（ADR）
- 为什么选择 DDD 分层架构？
- 为什么使用策略模式实现负载均衡？
- 为什么使用装饰器模式实现亲和性？
- 为什么使用 go-cache 而非 Redis 做亲和性缓存？

## 输出要求
- 使用中文撰写
- 所有图表使用 Mermaid 语法
- 专业术语保留英文（如 DDD、Repository、Entity）
- 文档应具备可交付质量，可直接用于团队内部技术评审
