// Package entity 定义了 API Instance 领域的核心实体、枚举和值对象。
//
// 该包属于 DDD 架构的领域层，包含 API 实例的聚合根 ApiInstanceEntity，
// 以及负载均衡类型、亲和性上下文等值对象。
package entity

import (
	"fmt"
	"time"

	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// ApiInstanceStatus 定义了 API 实例的生命周期状态。
//
// 实例状态决定了该实例是否能参与路由选择。
type ApiInstanceStatus string

const (
	ApiInstanceStatusActive     ApiInstanceStatus = "ACTIVE"     // 激活状态，可正常参与路由选择
	ApiInstanceStatusInactive   ApiInstanceStatus = "INACTIVE"   // 停用状态，暂时不参与路由
	ApiInstanceStatusDeprecated ApiInstanceStatus = "DEPRECATED" // 已弃用，即将下线
)

// ApiInstanceStatusFromCode 根据代码获取状态
func ApiInstanceStatusFromCode(code string) (ApiInstanceStatus, error) {
	switch code {
	case "ACTIVE":
		return ApiInstanceStatusActive, nil
	case "INACTIVE":
		return ApiInstanceStatusInactive, nil
	case "DEPRECATED":
		return ApiInstanceStatusDeprecated, nil
	default:
		return "", fmt.Errorf("未知的 API 实例状态代码: %s", code)
	}
}

// ApiType 定义了 API 实例的业务类型分类。
//
// 不同类型的 API 实例在同一项目下可以独立管理和路由。
type ApiType string

const (
	ApiTypeModel               ApiType = "MODEL"                // AI 模型服务（如 GPT、Claude）
	ApiTypePaymentGateway      ApiType = "PAYMENT_GATEWAY"      // 支付网关
	ApiTypeNotificationService ApiType = "NOTIFICATION_SERVICE" // 通知服务
	ApiTypeSmsService          ApiType = "SMS_SERVICE"          // 短信服务
	ApiTypeEmailService        ApiType = "EMAIL_SERVICE"        // 邮件服务
	ApiTypeFileStorage         ApiType = "FILE_STORAGE"         // 文件存储
	ApiTypeImageProcessing     ApiType = "IMAGE_PROCESSING"     // 图片处理
	ApiTypeOther               ApiType = "OTHER"                // 其他类型
)

// ApiTypeFromCode 根据代码获取类型
func ApiTypeFromCode(code string) (ApiType, error) {
	switch code {
	case "MODEL":
		return ApiTypeModel, nil
	case "PAYMENT_GATEWAY":
		return ApiTypePaymentGateway, nil
	case "NOTIFICATION_SERVICE":
		return ApiTypeNotificationService, nil
	case "SMS_SERVICE":
		return ApiTypeSmsService, nil
	case "EMAIL_SERVICE":
		return ApiTypeEmailService, nil
	case "FILE_STORAGE":
		return ApiTypeFileStorage, nil
	case "IMAGE_PROCESSING":
		return ApiTypeImageProcessing, nil
	case "OTHER":
		return ApiTypeOther, nil
	default:
		return "", fmt.Errorf("未知的 API 类型代码: %s", code)
	}
}

// LoadBalancingType 定义了可用的负载均衡策略类型。
//
// 每种策略对应 strategy 包中的一个具体实现。
type LoadBalancingType string

const (
	LoadBalancingTypeSmart            LoadBalancingType = "smart"              // 智能策略：综合评分（成功率×0.4 + 延迟×0.4 + 负载×0.2）
	LoadBalancingTypeRoundRobin       LoadBalancingType = "round_robin"        // 轮询策略：依次选择每个可用实例
	LoadBalancingTypeSuccessRateFirst LoadBalancingType = "success_rate_first" // 成功率优先：选择历史成功率最高的实例
	LoadBalancingTypeLatencyFirst     LoadBalancingType = "latency_first"      // 延迟优先：选择平均延迟最低的实例
)

// LoadBalancingTypeFromCode 根据代码获取类型
func LoadBalancingTypeFromCode(code string) (LoadBalancingType, error) {
	switch code {
	case "smart":
		return LoadBalancingTypeSmart, nil
	case "round_robin":
		return LoadBalancingTypeRoundRobin, nil
	case "success_rate_first":
		return LoadBalancingTypeSuccessRateFirst, nil
	case "latency_first":
		return LoadBalancingTypeLatencyFirst, nil
	default:
		return "", fmt.Errorf("未知的负载均衡策略类型: %s", code)
	}
}

// AffinityStrength 定义了亲和性绑定的强度级别。
//
// 强度决定了当绑定的实例不可用时的行为。
type AffinityStrength string

const (
	AffinityStrengthStrict    AffinityStrength = "strict"    // 严格模式：绑定实例不可用时直接报错
	AffinityStrengthPreferred AffinityStrength = "preferred" // 优先模式：绑定实例不可用时清除绑定并重新选择
	AffinityStrengthNone      AffinityStrength = "none"      // 无亲和性：不使用亲和性绑定
)

// AffinityContext 亲和性上下文值对象。
//
// 用于描述一次请求的亲和性需求，包括亲和性类型（如 "user"）、
// 亲和性键（如用户 ID）和绑定强度。
type AffinityContext struct {
	AffinityType string           `json:"affinity_type"`
	AffinityKey  string           `json:"affinity_key"`
	Strength     AffinityStrength `json:"strength"`
	ExpiresAt    *time.Time       `json:"expires_at"`
}

// GetBindingKey 构建亲和性绑定的唯一键
func (a *AffinityContext) GetBindingKey() string {
	return a.AffinityType + ":" + a.AffinityKey
}

// IsValid 检查亲和性是否有效
func (a *AffinityContext) IsValid() bool {
	return a != nil && a.AffinityType != "" && a.AffinityKey != ""
}

// IsExpired 检查是否已过期
func (a *AffinityContext) IsExpired() bool {
	return a.ExpiresAt != nil && time.Now().After(*a.ExpiresAt)
}

// AffinityBinding 亲和性绑定值对象。
//
// 记录某个亲和性键与特定 API 实例之间的绑定关系，
// 包含创建时间、过期时间、使用次数等生命周期信息。
// 绑定存储在 go-cache 本地缓存中，默认 60 分钟过期。
type AffinityBinding struct {
	InstanceID   string    `json:"instance_id"`
	CreateTime   time.Time `json:"create_time"`
	ExpireTime   time.Time `json:"expire_time"`
	UseCount     int       `json:"use_count"`
	LastUsedTime time.Time `json:"last_used_time"`
}

// NewAffinityBinding 创建亲和性绑定
func NewAffinityBinding(instanceID string) *AffinityBinding {
	now := time.Now()
	return &AffinityBinding{
		InstanceID:   instanceID,
		CreateTime:   now,
		ExpireTime:   now.Add(30 * time.Minute),
		UseCount:     1,
		LastUsedTime: now,
	}
}

// WithNewExpireTime 创建新的绑定对象，增加使用次数并更新过期时间
func (b *AffinityBinding) WithNewExpireTime(newExpireTime time.Time) *AffinityBinding {
	return &AffinityBinding{
		InstanceID:   b.InstanceID,
		CreateTime:   b.CreateTime,
		ExpireTime:   newExpireTime,
		UseCount:     b.UseCount + 1,
		LastUsedTime: time.Now(),
	}
}

// IsExpired 检查绑定是否已过期
func (b *AffinityBinding) IsExpired() bool {
	return time.Now().After(b.ExpireTime)
}

// ApiInstanceEntity 是 API Instance 领域的聚合根。
//
// 代表一个注册到 Gateway 的后端 API 实例，包含实例的基本信息、
// 路由参数（优先级、成本、权重）和扩展元数据。
// 对应数据库表 api_instance_registry。
type ApiInstanceEntity struct {
	ID            string                 `gorm:"column:id;primaryKey" json:"id"`             // 唯一标识符（UUID）
	ProjectID     string                 `gorm:"column:project_id" json:"project_id"`         // 所属项目 ID
	UserID        string                 `gorm:"column:user_id" json:"user_id"`               // 所属用户 ID（可选，用于用户级隔离）
	ApiIdentifier string                 `gorm:"column:api_identifier" json:"api_identifier"` // API 逻辑标识符（如 "gpt4o"）
	ApiType       ApiType                `gorm:"column:api_type" json:"api_type"`             // API 业务类型
	BusinessID    string                 `gorm:"column:business_id" json:"business_id"`       // 业务 ID（项目方内部标识）
	RoutingParams metricsEntity.JSONBMap `gorm:"column:routing_params;type:jsonb" json:"routing_params"` // 路由参数（priority/cost_per_unit/initial_weight）
	Status        ApiInstanceStatus      `gorm:"column:status" json:"status"`                 // 实例状态
	Metadata      metricsEntity.JSONBMap `gorm:"column:metadata;type:jsonb" json:"metadata"`  // 扩展元数据（provider/region 等）
	CreatedAt     time.Time              `gorm:"column:created_at;autoCreateTime" json:"created_at"` // 创建时间
	UpdatedAt     time.Time              `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"` // 更新时间
}

// TableName 指定表名
func (ApiInstanceEntity) TableName() string {
	return "api_instance_registry"
}

// Activate 激活API实例
func (a *ApiInstanceEntity) Activate() {
	a.Status = ApiInstanceStatusActive
}

// Deactivate 停用API实例
func (a *ApiInstanceEntity) Deactivate() {
	a.Status = ApiInstanceStatusInactive
}

// Deprecate 标记为已弃用
func (a *ApiInstanceEntity) Deprecate() {
	a.Status = ApiInstanceStatusDeprecated
}

// IsAvailable 检查API实例是否可用
func (a *ApiInstanceEntity) IsAvailable() bool {
	return a.Status == ApiInstanceStatusActive
}

// GetPriority 获取优先级
func (a *ApiInstanceEntity) GetPriority() int {
	if a.RoutingParams != nil {
		if v, ok := a.RoutingParams["priority"]; ok {
			if priority, ok := v.(float64); ok {
				return int(priority)
			}
		}
	}
	return 0
}

// GetCostPerUnit 获取单位成本
func (a *ApiInstanceEntity) GetCostPerUnit() float64 {
	if a.RoutingParams != nil {
		if v, ok := a.RoutingParams["cost_per_unit"]; ok {
			if cost, ok := v.(float64); ok {
				return cost
			}
		}
	}
	return 0.0
}

// GetInitialWeight 获取初始权重
func (a *ApiInstanceEntity) GetInitialWeight() int {
	if a.RoutingParams != nil {
		if v, ok := a.RoutingParams["initial_weight"]; ok {
			if weight, ok := v.(float64); ok {
				return int(weight)
			}
		}
	}
	return 1
}
