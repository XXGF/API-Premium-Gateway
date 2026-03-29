# Prompt 3：代码注释补充

## 角色
你是一位资深的 Go 语言工程师，精通 Go 官方注释规范（godoc 风格）和 DDD 架构。

## 目标
请阅读 `go-gateway/` 目录下的所有 Go 源代码文件，为每个文件补充完善且规范的中文注释。**直接修改源文件**，不要只输出注释内容。

## 注释规范要求

### 1. 包注释（Package Comment）
每个包的第一个文件顶部必须有包注释，格式：
```go
// Package xxx 提供了 xxx 功能。
//
// 该包属于 DDD 架构的 xxx 层，负责 xxx。
package xxx
```

### 2. 导出类型注释（Exported Type）
每个导出的 struct、interface 必须有注释，格式：
```go
// XxxEntity 是 xxx 领域的核心实体/聚合根。
//
// 该实体包含 xxx 信息，对应数据库表 xxx。
// 主要职责：xxx。
type XxxEntity struct {
```

### 3. 导出字段注释
struct 中每个导出字段必须有行内注释或上方注释，说明业务含义：
```go
type XxxEntity struct {
    ID        string `...` // 唯一标识符（UUID）
    Status    string `...` // 当前状态，可选值：ACTIVE、INACTIVE、DEPRECATED
}
```

### 4. 导出函数/方法注释
每个导出的函数和方法必须有注释，格式：
```go
// NewXxxService 创建 xxx 服务实例。
//
// 参数：
//   - repo: xxx 仓储接口实现
//   - cache: 缓存服务（可选）
//
// 返回新创建的服务实例。
func NewXxxService(repo XxxRepository) *XxxService {
```

对于复杂的业务方法，还需要说明：
- 业务逻辑概述
- 关键算法说明
- 可能返回的错误类型

### 5. 接口注释
接口的每个方法必须有注释，说明契约：
```go
// XxxRepository 定义了 xxx 领域的持久化接口。
type XxxRepository interface {
    // Insert 插入一条新记录。
    // 如果 ID 为空，将自动生成 UUID。
    Insert(entity *XxxEntity) error
}
```

### 6. 枚举/常量注释
每个常量组必须有注释，每个常量值说明含义：
```go
// ApiInstanceStatus 定义了 API 实例的生命周期状态。
type ApiInstanceStatus string

const (
    ApiInstanceStatusActive     ApiInstanceStatus = "ACTIVE"     // 激活状态，可正常参与路由选择
    ApiInstanceStatusInactive   ApiInstanceStatus = "INACTIVE"   // 停用状态，暂时不参与路由
    ApiInstanceStatusDeprecated ApiInstanceStatus = "DEPRECATED" // 已弃用，即将下线
)
```

### 7. 策略模式注释
对于策略模式相关代码，需要额外说明：
- 策略接口的设计意图
- 每种策略的算法原理和适用场景
- 工厂的注册机制
- 装饰器的增强逻辑

### 8. 关键算法注释
对于以下复杂逻辑，需要添加详细的行内注释：
- SmartStrategy 的综合评分算法（权重计算、得分公式）
- 熔断判断逻辑（阈值、条件）
- 亲和性绑定的生命周期管理
- 降级链（Fallback Chain）的执行逻辑

## 处理顺序
请按以下顺序逐个文件处理：
1. `internal/domain/` — 领域层（最核心，注释最详细）
2. `internal/application/` — 应用层
3. `internal/interfaces/` — 接口层
4. `internal/infrastructure/` — 基础设施层
5. `cmd/server/main.go` — 入口文件

## 注意事项
- 注释语言：中文（专业术语如 DDD、Repository、Entity 等保留英文）
- 遵循 Go 官方 godoc 注释风格（注释以被注释对象的名称开头）
- 不要修改任何业务逻辑代码，只添加/完善注释
- 不要添加冗余注释（如 `i++ // i 自增 1` 这种）
- 注释应体现业务语义，而非代码翻译
- 对于已有注释，如果不够完善则补充，如果已经足够好则保留
