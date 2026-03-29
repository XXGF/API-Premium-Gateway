package service

import (
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apikey/service"
)

// AuthenticationAppService 认证应用服务
type AuthenticationAppService struct {
	apiKeyDomainService *service.ApiKeyDomainService
}

// NewAuthenticationAppService 创建认证应用服务
func NewAuthenticationAppService(apiKeyService *service.ApiKeyDomainService) *AuthenticationAppService {
	return &AuthenticationAppService{apiKeyDomainService: apiKeyService}
}

// IsValidApiKey 校验 API Key 是否有效
func (s *AuthenticationAppService) IsValidApiKey(apiKeyValue string) bool {
	return s.apiKeyDomainService.IsValidApiKey(apiKeyValue)
}
