// Package middleware 实现了 HTTP 中间件。
//
// 该包属于 DDD 架构的接口层，提供：
//   - ApiKeyAuthMiddleware：API Key 认证中间件
//   - GlobalExceptionHandler：全局异常捕获中间件
//   - CORSMiddleware：跨域资源共享中间件
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	appService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/service"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/response"
)

// API Key 认证相关常量
const (
	apiKeyHeader    = "X-API-Key"  // 请求头中的 API Key 字段名
	apiKeyParam     = "apiKey"     // 查询参数中的 API Key 字段名
	projectIDCtxKey = "projectId"  // Gin 上下文中存储项目 ID 的键
)

// ApiKeyAuthMiddleware API Key 认证中间件。
//
// 支持两种传递方式：
//   - 请求头：X-API-Key: <key>
//   - 查询参数：?apiKey=<key>
//
// 认证失败时返回 401 并终止请求链。
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

// GlobalExceptionHandler 全局异常处理中间件。
//
// 捕获 Handler 中未处理的 panic，记录错误日志并返回 500 响应，
// 防止服务崩溃。
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

// CORSMiddleware 跨域资源共享中间件。
//
// 允许所有来源的跨域请求，支持 GET/POST/PUT/DELETE/OPTIONS 方法。
// 对于 OPTIONS 预检请求，直接返回 204。
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

// maskApiKey 对 API Key 进行脱敏处理。
//
// 保留前 4 位和后 4 位，中间用 **** 替代。
// 用于日志输出，防止 Key 泄露。
func maskApiKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}
