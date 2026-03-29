package entity

import (
	"fmt"
	"time"

	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// ApiInstanceStatus API实例状态
type ApiInstanceStatus string

const (
	ApiInstanceStatusActive     ApiInstanceStatus = "ACTIVE"
	ApiInstanceStatusInactive   ApiInstanceStatus = "INACTIVE"
	ApiInstanceStatusDeprecated ApiInstanceStatus = "DEPRECATED"
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

// ApiType API类型
type ApiType string

const (
	ApiTypeModel               ApiType = "MODEL"
	ApiTypePaymentGateway      ApiType = "PAYMENT_GATEWAY"
	ApiTypeNotificationService ApiType = "NOTIFICATION_SERVICE"
	ApiTypeSmsService          ApiType = "SMS_SERVICE"
	ApiTypeEmailService        ApiType = "EMAIL_SERVICE"
	ApiTypeFileStorage         ApiType = "FILE_STORAGE"
	ApiTypeImageProcessing     ApiType = "IMAGE_PROCESSING"
	ApiTypeOther               ApiType = "OTHER"
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

// LoadBalancingType 负载均衡策略类型
type LoadBalancingType string

const (
	LoadBalancingTypeSmart            LoadBalancingType = "smart"
	LoadBalancingTypeRoundRobin       LoadBalancingType = "round_robin"
	LoadBalancingTypeSuccessRateFirst LoadBalancingType = "success_rate_first"
	LoadBalancingTypeLatencyFirst     LoadBalancingType = "latency_first"
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

// AffinityStrength 亲和性强度
type AffinityStrength string

const (
	AffinityStrengthStrict    AffinityStrength = "strict"
	AffinityStrengthPreferred AffinityStrength = "preferred"
	AffinityStrengthNone      AffinityStrength = "none"
)

// AffinityContext 亲和性上下文
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

// AffinityBinding 亲和性绑定对象
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

// ApiInstanceEntity API实例领域实体
type ApiInstanceEntity struct {
	ID            string                   `gorm:"column:id;primaryKey" json:"id"`
	ProjectID     string                   `gorm:"column:project_id" json:"project_id"`
	UserID        string                   `gorm:"column:user_id" json:"user_id"`
	ApiIdentifier string                   `gorm:"column:api_identifier" json:"api_identifier"`
	ApiType       ApiType                  `gorm:"column:api_type" json:"api_type"`
	BusinessID    string                   `gorm:"column:business_id" json:"business_id"`
	RoutingParams metricsEntity.JSONBMap   `gorm:"column:routing_params;type:jsonb" json:"routing_params"`
	Status        ApiInstanceStatus        `gorm:"column:status" json:"status"`
	Metadata      metricsEntity.JSONBMap   `gorm:"column:metadata;type:jsonb" json:"metadata"`
	CreatedAt     time.Time                `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time                `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
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
