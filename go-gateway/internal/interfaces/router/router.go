package router

import (
	"github.com/gin-gonic/gin"

	appService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/service"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/handler"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/middleware"
)

// SetupRouter 设置路由
func SetupRouter(
	authService *appService.AuthenticationAppService,
	selectionAppService *appService.SelectionAppService,
	apiInstanceAppService *appService.ApiInstanceAppService,
) *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.GlobalExceptionHandler())

	// 创建 Handler
	gatewayHandler := handler.NewGatewayHandler(selectionAppService)
	apiInstanceHandler := handler.NewApiInstanceHandler(apiInstanceAppService)

	// API 路由组
	api := r.Group("/api")
	{
		// 健康检查（无需认证）
		api.GET("/health", handler.HealthCheck)

		// 需要 API Key 认证的路由
		authenticated := api.Group("")
		authenticated.Use(middleware.ApiKeyAuthMiddleware(authService))
		{
			// 外部接口 - 网关核心
			external := authenticated.Group("/external")
			{
				// 网关选择和上报
				gateway := external.Group("/gateway")
				{
					gateway.POST("/select", gatewayHandler.SelectInstance)
					gateway.POST("/report", gatewayHandler.ReportResult)
				}

				// API 实例管理
				instances := external.Group("/api-instances")
				{
					instances.POST("/:projectId", apiInstanceHandler.CreateApiInstance)
					instances.POST("/:projectId/batch", apiInstanceHandler.BatchCreateApiInstances)
					instances.GET("/:projectId", apiInstanceHandler.GetApiInstancesByProjectId)
					instances.GET("/detail/:id", apiInstanceHandler.GetApiInstanceById)
					instances.PUT("/:projectId/:apiType/:businessId", apiInstanceHandler.UpdateApiInstance)
					instances.DELETE("/:projectId/:apiType/:businessId", apiInstanceHandler.DeleteApiInstance)
					instances.DELETE("/:projectId/batch", apiInstanceHandler.BatchDeleteApiInstances)
					instances.PUT("/:projectId/:apiType/:businessId/activate", apiInstanceHandler.ActivateApiInstance)
					instances.PUT("/:projectId/:apiType/:businessId/deactivate", apiInstanceHandler.DeactivateApiInstance)
				}
			}

			// 管理接口
			admin := authenticated.Group("/admin")
			{
				admin.GET("/api-instances", apiInstanceHandler.GetAllInstances)
			}
		}
	}

	return r
}
