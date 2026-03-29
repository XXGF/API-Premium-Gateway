package repository

import (
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apikey/entity"
)

// ApiKeyRepository API Key 仓储接口
type ApiKeyRepository interface {
	// Insert 插入 API Key
	Insert(apiKey *entity.ApiKeyEntity) error
	// SelectByID 根据ID查询
	SelectByID(id string) (*entity.ApiKeyEntity, error)
	// SelectByApiKeyValue 根据 API Key 值查询
	SelectByApiKeyValue(apiKeyValue string) (*entity.ApiKeyEntity, error)
	// SelectAll 查询所有（按创建时间降序）
	SelectAll() ([]*entity.ApiKeyEntity, error)
	// Update 更新 API Key
	Update(apiKey *entity.ApiKeyEntity) error
	// UpdateStatus 更新状态
	UpdateStatus(id string, status entity.ApiKeyStatus) error
	// UpdateDescription 更新描述和过期时间
	UpdateFields(id string, fields map[string]interface{}) error
	// DeleteByID 根据ID删除
	DeleteByID(id string) (int64, error)
	// CountByApiKeyValue 根据 API Key 值统计数量
	CountByApiKeyValue(apiKeyValue string) (int64, error)
}
