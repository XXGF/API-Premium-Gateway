// Package service 实现了 ApiKey 领域的核心业务逻辑。
//
// 该包属于 DDD 架构的领域层，包含 ApiKeyDomainService，
// 负责 API Key 的生成、校验、生命周期管理等核心业务逻辑。
package service

import (
	"crypto/rand"
	"math/big"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apikey/entity"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apikey/repository"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/infrastructure/exception"
)

// API Key 生成的配置常量
const (
	apiKeyPrefix = "gw_"  // Key 值前缀，用于标识 Gateway 生成的 Key
	apiKeyLength = 32     // Key 随机部分的字符长度
	apiKeyChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789" // Key 随机字符集
	maxAttempts  = 10     // 生成唯一 Key 的最大尝试次数
)

// ApiKeyDomainService API Key 领域服务。
//
// 负责 API Key 的生成、校验、CRUD 等核心业务逻辑。
// Key 生成规则："gw_" 前缀 + 32 位加密安全随机字符，保证唯一性。
type ApiKeyDomainService struct {
	apiKeyRepository repository.ApiKeyRepository
}

// NewApiKeyDomainService 创建 API Key 领域服务
func NewApiKeyDomainService(repo repository.ApiKeyRepository) *ApiKeyDomainService {
	return &ApiKeyDomainService{apiKeyRepository: repo}
}

// GenerateApiKey 生成新的 API Key。
//
// 自动生成唯一的 Key 值（最多尝试 10 次），初始状态为 UNUSED。
func (s *ApiKeyDomainService) GenerateApiKey(description string, expiresAt *time.Time) (*entity.ApiKeyEntity, error) {
	log.Info().Str("description", description).Msg("生成 API Key")

	apiKeyValue, err := s.generateUniqueApiKeyValue()
	if err != nil {
		return nil, err
	}

	apiKey := entity.NewApiKeyEntityWithExpiry(apiKeyValue, description, expiresAt)
	if err := s.apiKeyRepository.Insert(apiKey); err != nil {
		return nil, err
	}

	log.Info().Str("id", apiKey.ID).Msg("成功生成 API Key")
	return apiKey, nil
}

// GetById 根据ID获取 API Key
func (s *ApiKeyDomainService) GetById(id string) (*entity.ApiKeyEntity, error) {
	apiKey, err := s.apiKeyRepository.SelectByID(id)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, exception.NewEntityNotFoundError("API Key 不存在，ID: " + id)
	}
	return apiKey, nil
}

// FindAll 获取所有 API Key
func (s *ApiKeyDomainService) FindAll() ([]*entity.ApiKeyEntity, error) {
	return s.apiKeyRepository.SelectAll()
}

// UpdateApiKey 修改 API Key
func (s *ApiKeyDomainService) UpdateApiKey(id string, description *string, expiresAt *time.Time) (bool, error) {
	fields := make(map[string]interface{})
	if description != nil {
		fields["description"] = *description
	}
	if expiresAt != nil {
		fields["expires_at"] = *expiresAt
	}
	if len(fields) == 0 {
		return false, nil
	}

	if err := s.apiKeyRepository.UpdateFields(id, fields); err != nil {
		return false, err
	}
	log.Info().Str("id", id).Msg("修改 API Key 成功")
	return true, nil
}

// UpdateStatus 修改 API Key 状态
func (s *ApiKeyDomainService) UpdateStatus(id string, status entity.ApiKeyStatus) (bool, error) {
	if err := s.apiKeyRepository.UpdateStatus(id, status); err != nil {
		return false, err
	}
	log.Info().Str("id", id).Str("status", string(status)).Msg("修改 API Key 状态成功")
	return true, nil
}

// DeleteById 删除 API Key
func (s *ApiKeyDomainService) DeleteById(id string) (bool, error) {
	deleted, err := s.apiKeyRepository.DeleteByID(id)
	if err != nil {
		return false, err
	}
	if deleted > 0 {
		log.Info().Str("id", id).Msg("删除 API Key 成功")
	}
	return deleted > 0, nil
}

// IsUsable 检查 API Key 是否可用
func (s *ApiKeyDomainService) IsUsable(apiKeyValue string) bool {
	apiKey, err := s.apiKeyRepository.SelectByApiKeyValue(apiKeyValue)
	if err != nil || apiKey == nil {
		return false
	}
	return apiKey.IsUsable()
}

// IsValidApiKey 校验 API Key 是否有效（用于认证中间件）。
//
// 检查 Key 是否存在且状态可用（ACTIVE 或 UNUSED）。
func (s *ApiKeyDomainService) IsValidApiKey(apiKeyValue string) bool {
	return s.IsUsable(apiKeyValue)
}

// UseApiKey 使用 API Key：修改状态为激活
func (s *ApiKeyDomainService) UseApiKey(apiKeyValue string) error {
	return s.apiKeyRepository.UpdateFields("", map[string]interface{}{
		"status":        entity.ApiKeyStatusActive,
		"api_key_value": apiKeyValue,
	})
}

// CreateApiKey 创建 API Key
func (s *ApiKeyDomainService) CreateApiKey(defaultApiKey string) error {
	apiKey := entity.NewApiKeyEntity(defaultApiKey, "默认的")
	return s.apiKeyRepository.Insert(apiKey)
}

// ExistApiKey 检查 API Key 是否存在
func (s *ApiKeyDomainService) ExistApiKey(apiKeyValue string) bool {
	apiKey, err := s.apiKeyRepository.SelectByApiKeyValue(apiKeyValue)
	return err == nil && apiKey != nil
}

// generateUniqueApiKeyValue 生成唯一的 API Key 值。
//
// 最多尝试 maxAttempts 次，每次生成后检查数据库中是否已存在。
// 如果超过最大尝试次数仍未生成唯一值，返回错误。
func (s *ApiKeyDomainService) generateUniqueApiKeyValue() (string, error) {
	for i := 0; i < maxAttempts; i++ {
		apiKeyValue := s.generateApiKeyValue()
		count, err := s.apiKeyRepository.CountByApiKeyValue(apiKeyValue)
		if err != nil {
			return "", err
		}
		if count == 0 {
			return apiKeyValue, nil
		}
	}
	return "", exception.ApiKeyGenerationFailedError("尝试次数超过最大限制")
}

// generateApiKeyValue 生成 API Key 值。
//
// 格式："gw_" + 32 位加密安全随机字符（使用 crypto/rand）。
func (s *ApiKeyDomainService) generateApiKeyValue() string {
	result := apiKeyPrefix
	chars := []byte(apiKeyChars)
	for i := 0; i < apiKeyLength; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		result += string(chars[n.Int64()])
	}
	return result
}
