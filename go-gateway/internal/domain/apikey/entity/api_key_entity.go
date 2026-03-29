package entity

import (
	"fmt"
	"time"
)

// ApiKeyStatus API Key 状态
type ApiKeyStatus string

const (
	ApiKeyStatusActive  ApiKeyStatus = "ACTIVE"
	ApiKeyStatusRevoked ApiKeyStatus = "REVOKED"
	ApiKeyStatusExpired ApiKeyStatus = "EXPIRED"
	ApiKeyStatusUnused  ApiKeyStatus = "UNUSED"
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

// ApiKeyEntity API Key 领域实体
type ApiKeyEntity struct {
	ID          string       `gorm:"column:id;primaryKey" json:"id"`
	ApiKeyValue string       `gorm:"column:api_key_value" json:"api_key_value"`
	Description string       `gorm:"column:description" json:"description"`
	Status      ApiKeyStatus `gorm:"column:status" json:"status"`
	IssuedAt    time.Time    `gorm:"column:issued_at;autoCreateTime" json:"issued_at"`
	ExpiresAt   *time.Time   `gorm:"column:expires_at" json:"expires_at"`
	LastUsedAt  *time.Time   `gorm:"column:last_used_at" json:"last_used_at"`
	CreatedAt   time.Time    `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
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

// IsUsable 检查 API Key 是否可用
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

// GetRemainingDays 获取剩余有效天数
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
