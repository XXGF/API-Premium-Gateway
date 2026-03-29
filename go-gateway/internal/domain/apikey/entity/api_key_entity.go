// Package entity 定义了 ApiKey 领域的核心实体和枚举。
//
// 该包属于 DDD 架构的领域层，包含 ApiKeyEntity 聚合根，
// 用于管理 API Key 的生命周期（创建、激活、撤销、过期）。
package entity

import (
	"fmt"
	"time"
)

// ApiKeyStatus 定义了 API Key 的生命周期状态。
type ApiKeyStatus string

const (
	ApiKeyStatusActive  ApiKeyStatus = "ACTIVE"  // 激活状态，可正常使用
	ApiKeyStatusRevoked ApiKeyStatus = "REVOKED" // 已撤销，永久失效
	ApiKeyStatusExpired ApiKeyStatus = "EXPIRED" // 已过期，超过有效期
	ApiKeyStatusUnused  ApiKeyStatus = "UNUSED"  // 未使用，刚创建尚未首次使用
)

// ApiKeyStatusFromCode 根据代码获取状态
func ApiKeyStatusFromCode(code string) (ApiKeyStatus, error) {
	switch code {
	case "ACTIVE":
		return ApiKeyStatusActive, nil
	case "REVOKED":
		return ApiKeyStatusRevoked, nil
	case "EXPIRED":
		return ApiKeyStatusExpired, nil
	case "UNUSED":
		return ApiKeyStatusUnused, nil
	default:
		return "", fmt.Errorf("未知的 API Key 状态代码: %s", code)
	}
}

// IsUsable 检查是否为可用状态
func (s ApiKeyStatus) IsUsable() bool {
	return s == ApiKeyStatusActive || s == ApiKeyStatusUnused
}

// ApiKeyEntity 是 ApiKey 领域的聚合根。
//
// 代表一个独立管理的 API Key，用于接口认证。
// Key 值以 "gw_" 前缀 + 32 位随机字符生成。
// 对应数据库表 api_keys。
type ApiKeyEntity struct {
	ID          string       `gorm:"column:id;primaryKey" json:"id"`                       // 唯一标识符（UUID）
	ApiKeyValue string       `gorm:"column:api_key_value" json:"api_key_value"`             // Key 值（gw_ 前缀 + 32 位随机字符）
	Description string       `gorm:"column:description" json:"description"`                // 描述信息
	Status      ApiKeyStatus `gorm:"column:status" json:"status"`                           // 当前状态
	IssuedAt    time.Time    `gorm:"column:issued_at;autoCreateTime" json:"issued_at"`      // 颁发时间
	ExpiresAt   *time.Time   `gorm:"column:expires_at" json:"expires_at"`                   // 过期时间（nil 表示永不过期）
	LastUsedAt  *time.Time   `gorm:"column:last_used_at" json:"last_used_at"`               // 最后使用时间
	CreatedAt   time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`    // 创建时间
	UpdatedAt   time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`    // 更新时间
}

// TableName 指定表名
func (ApiKeyEntity) TableName() string {
	return "api_keys"
}

// NewApiKeyEntity 创建 API Key 实体
func NewApiKeyEntity(apiKeyValue, description string) *ApiKeyEntity {
	return &ApiKeyEntity{
		ApiKeyValue: apiKeyValue,
		Description: description,
		Status:      ApiKeyStatusUnused,
	}
}

// NewApiKeyEntityWithExpiry 创建带过期时间的 API Key 实体
func NewApiKeyEntityWithExpiry(apiKeyValue, description string, expiresAt *time.Time) *ApiKeyEntity {
	entity := NewApiKeyEntity(apiKeyValue, description)
	entity.ExpiresAt = expiresAt
	return entity
}

// Activate 激活 API Key
func (a *ApiKeyEntity) Activate() error {
	if a.Status == ApiKeyStatusRevoked {
		return fmt.Errorf("已撤销的 API Key 无法激活")
	}
	a.Status = ApiKeyStatusActive
	return nil
}

// Revoke 撤销 API Key
func (a *ApiKeyEntity) Revoke() {
	a.Status = ApiKeyStatusRevoked
}

// MarkExpired 标记为已过期
func (a *ApiKeyEntity) MarkExpired() {
	a.Status = ApiKeyStatusExpired
}

// IsUsable 检查 API Key 是否可用。
//
// 可用条件：状态为 ACTIVE 或 UNUSED，且未超过过期时间。
// 如果检测到已过期，会自动将状态更新为 EXPIRED。
func (a *ApiKeyEntity) IsUsable() bool {
	if !a.Status.IsUsable() {
		return false
	}
	// 检查是否过期
	if a.ExpiresAt != nil && time.Now().After(*a.ExpiresAt) {
		a.MarkExpired()
		return false
	}
	return true
}

// IsExpired 检查 API Key 是否已过期
func (a *ApiKeyEntity) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// UpdateLastUsedTime 更新最后使用时间
func (a *ApiKeyEntity) UpdateLastUsedTime() {
	now := time.Now()
	a.LastUsedAt = &now
	if a.Status == ApiKeyStatusUnused {
		a.Status = ApiKeyStatusActive
	}
}

// UpdateDescription 更新描述
func (a *ApiKeyEntity) UpdateDescription(description string) {
	a.Description = description
}

// GetRemainingDays 获取剩余有效天数。
//
// 返回值：
//   - nil：永不过期
//   - 0：已过期
//   - >0：剩余天数
func (a *ApiKeyEntity) GetRemainingDays() *int64 {
	if a.ExpiresAt == nil {
		return nil // 永不过期
	}
	now := time.Now()
	if now.After(*a.ExpiresAt) {
		days := int64(0)
		return &days
	}
	days := int64(a.ExpiresAt.Sub(now).Hours() / 24)
	return &days
}
