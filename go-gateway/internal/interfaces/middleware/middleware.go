package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	appService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/service"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/response"
)

const (
	apiKeyHeader    = "X-API-Key"
	apiKeyParam     = "apiKey"
	projectIDCtxKey = "projectId"
)

// ApiKeyAuthMiddleware API Key 认证中间件
func ApiKeyAuthMiddleware(authService *appService.AuthenticationAppService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 或 Query 参数中获取 API Key
		apiKey := c.GetHeader(apiKeyHeader)
		if apiKey == "" {
			apiKey = c.Query(apiKeyParam)
		}

		if apiKey == "" {
			log.Warn().Str("path", c.Request.URL.Path).Msg("缺少 API Key")
			c.JSON(http.StatusUnauthorized, response.Fail("缺少 API Key，请在请求头 X-API-Key 或查询参数 apiKey 中提供"))
			c.Abort()
			return
		}

		// 校验 API Key
		if !authService.IsValidApiKey(apiKey) {
			log.Warn().Str("apiKey", maskApiKey(apiKey)).Msg("无效的 API Key")
			c.JSON(http.StatusUnauthorized, response.Fail("无效的 API Key"))
			c.Abort()
			return
		}

		log.Debug().Str("apiKey", maskApiKey(apiKey)).Msg("API Key 认证通过")
		c.Next()
	}
}

// GlobalExceptionHandler 全局异常处理中间件
func GlobalExceptionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Interface("error", err).Str("path", c.Request.URL.Path).Msg("服务器内部错误")
				c.JSON(http.StatusInternalServerError, response.Fail("服务器内部错误"))
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// maskApiKey 脱敏 API Key
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}
