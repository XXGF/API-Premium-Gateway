# Prompt 2：API 接口文档

## 角色
你是一位资深的 API 设计师，精通 RESTful API 设计规范和 OpenAPI 文档标准。

## 目标
请阅读 `go-gateway/` 目录下的所有 Go 源代码，生成一份完善且专业的中文 API 接口文档（Markdown 格式）。

## 阅读范围
请重点阅读以下文件：
1. `internal/interfaces/router/router.go` — 路由定义（所有 API 路径和分组）
2. `internal/interfaces/handler/gateway_handler.go` — 网关核心接口
3. `internal/interfaces/handler/api_instance_handler.go` — API 实例管理接口
4. `internal/interfaces/middleware/middleware.go` — 认证中间件（了解鉴权方式）
5. `internal/interfaces/request/gateway_request.go` — 网关请求模型
6. `internal/interfaces/request/api_instance/api_instance_request.go` — 实例管理请求模型
7. `internal/interfaces/response/result.go` — 统一响应模型
8. `internal/application/dto/api_instance_dto.go` — 响应数据模型
9. `internal/infrastructure/exception/errors.go` — 错误码定义
10. `internal/domain/apiinstance/entity/api_instance_entity.go` — 了解枚举值（ApiType、Status 等）

## 文档结构要求

### 1. 文档概述
- API 基础信息（Base URL、版本、协议）
- 全局约定（请求/响应格式、时间格式、编码）

### 2. 认证方式
- 详细说明 API Key 认证机制
- 认证头格式（X-API-Key Header 或 apiKey Query 参数）
- 认证失败的响应示例

### 3. 统一响应格式
- 描述 Result 结构体的字段含义
- 成功响应示例
- 失败响应示例（含不同错误码）

### 4. 接口列表总览
- 用表格列出所有接口（方法、路径、描述、是否需要认证）

### 5. 接口详细文档
对每个接口，请提供以下信息：
- **接口名称**
- **请求方法和路径**
- **功能描述**
- **请求参数**（Path 参数、Query 参数、Body 参数，用表格列出字段名、类型、是否必填、描述）
- **请求示例**（完整的 curl 命令）
- **响应参数**（用表格列出字段名、类型、描述）
- **响应示例**（成功和失败各一个 JSON 示例）
- **错误码说明**（该接口可能返回的特定错误）

请按以下分组组织接口：

#### 5.1 健康检查
- GET /api/health

#### 5.2 网关核心接口（Gateway）
- POST /api/external/gateway/select — 选择最佳 API 实例
- POST /api/external/gateway/report — 上报调用结果

#### 5.3 API 实例管理接口
- POST /api/external/api-instances/:projectId — 创建 API 实例
- POST /api/external/api-instances/:projectId/batch — 批量创建
- GET /api/external/api-instances/:projectId — 获取项目下实例列表
- GET /api/external/api-instances/detail/:id — 获取实例详情
- PUT /api/external/api-instances/:projectId/:apiType/:businessId — 更新实例
- DELETE /api/external/api-instances/:projectId/:apiType/:businessId — 删除实例
- DELETE /api/external/api-instances/:projectId/batch — 批量删除
- PUT .../activate — 激活实例
- PUT .../deactivate — 停用实例

#### 5.4 管理接口（Admin）
- GET /api/admin/api-instances — 获取所有实例

### 6. 枚举值参考
- ApiType 所有可选值及含义
- ApiInstanceStatus 所有可选值及含义
- LoadBalancingType 所有可选值及含义
- AffinityStrength 所有可选值及含义
- GatewayStatus 所有可选值及含义

### 7. 典型使用流程
- 用时序图（Mermaid）展示完整的 SDK 集成流程：
  1. 注册 API 实例
  2. 请求选择最佳实例
  3. 上游服务执行实际调用
  4. 上报调用结果

## 输出要求
- 使用中文撰写，字段名保留英文
- curl 示例使用 localhost:8081 作为 Base URL
- JSON 示例需格式化且包含合理的示例数据
- 文档应具备可交付质量，可直接提供给前端/SDK 开发者使用
