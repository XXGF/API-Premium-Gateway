package persistence

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apikey/entity"
)

// ApiKeyRepositoryImpl API Key 仓储实现
type ApiKeyRepositoryImpl struct {
	db *gorm.DB
}

// NewApiKeyRepositoryImpl 创建 API Key 仓储实现
func NewApiKeyRepositoryImpl(db *gorm.DB) *ApiKeyRepositoryImpl {
	return &ApiKeyRepositoryImpl{db: db}
}

func (r *ApiKeyRepositoryImpl) Insert(apiKey *entity.ApiKeyEntity) error {
	if apiKey.ID == "" {
		apiKey.ID = uuid.New().String()
	}
	return r.db.Create(apiKey).Error
}

func (r *ApiKeyRepositoryImpl) SelectByID(id string) (*entity.ApiKeyEntity, error) {
	var apiKey entity.ApiKeyEntity
	result := r.db.Where("id = ?", id).First(&apiKey)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &apiKey, nil
}

func (r *ApiKeyRepositoryImpl) SelectByApiKeyValue(apiKeyValue string) (*entity.ApiKeyEntity, error) {
	var apiKey entity.ApiKeyEntity
	result := r.db.Where("api_key_value = ?", apiKeyValue).First(&apiKey)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &apiKey, nil
}

func (r *ApiKeyRepositoryImpl) SelectAll() ([]*entity.ApiKeyEntity, error) {
	var apiKeys []*entity.ApiKeyEntity
	result := r.db.Order("created_at DESC").Find(&apiKeys)
	return apiKeys, result.Error
}

func (r *ApiKeyRepositoryImpl) Update(apiKey *entity.ApiKeyEntity) error {
	return r.db.Save(apiKey).Error
}

func (r *ApiKeyRepositoryImpl) UpdateStatus(id string, status entity.ApiKeyStatus) error {
	return r.db.Model(&entity.ApiKeyEntity{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ApiKeyRepositoryImpl) UpdateFields(id string, fields map[string]interface{}) error {
	query := r.db.Model(&entity.ApiKeyEntity{})
	if id != "" {
		query = query.Where("id = ?", id)
	}
	// 如果 fields 中包含 api_key_value，用它作为查询条件
	if apiKeyValue, ok := fields["api_key_value"]; ok {
		query = query.Where("api_key_value = ?", apiKeyValue)
		delete(fields, "api_key_value")
	}
	return query.Updates(fields).Error
}

func (r *ApiKeyRepositoryImpl) DeleteByID(id string) (int64, error) {
	result := r.db.Where("id = ?", id).Delete(&entity.ApiKeyEntity{})
	return result.RowsAffected, result.Error
}

func (r *ApiKeyRepositoryImpl) CountByApiKeyValue(apiKeyValue string) (int64, error) {
	var count int64
	result := r.db.Model(&entity.ApiKeyEntity{}).Where("api_key_value = ?", apiKeyValue).Count(&count)
	return count, result.Error
}
