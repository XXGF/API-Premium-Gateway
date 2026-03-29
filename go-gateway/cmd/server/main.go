package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	// 应用层
	appService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/service"

	// 领域层 - 服务
	apiInstanceService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/service"
	apiInstanceStrategy "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/strategy"
	apiKeyService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apikey/service"
	metricsService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/service"
	projectService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/project/service"

	// 基础设施层
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/infrastructure/config"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/infrastructure/persistence"

	// 接口层
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/router"
)

func main() {
	// 1. 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Msg("========================================")
	log.Info().Msg("  API Premium Gateway (Go) 启动中...")
	log.Info().Msg("========================================")

	// 2. 加载配置
	configPath := "configs/config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("加载配置失败")
	}
	log.Info().Int("port", cfg.Server.Port).Str("contextPath", cfg.Server.ContextPath).Msg("配置加载成功")

	// 3. 初始化数据库
	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("初始化数据库失败")
	}
	log.Info().Str("host", cfg.Database.Host).Int("port", cfg.Database.Port).Str("dbname", cfg.Database.DBName).Msg("数据库连接成功")

	// 4. 初始化仓储层（基础设施层）
	projectRepo := persistence.NewProjectRepositoryImpl(db)
	apiKeyRepo := persistence.NewApiKeyRepositoryImpl(db)
	metricsRepo := persistence.NewMetricsRepositoryImpl(db)
	apiInstanceRepo := persistence.NewApiInstanceRepositoryImpl(db)

	// 5. 初始化领域层
	// 5.1 Project 领域服务
	projectDomainService := projectService.NewProjectDomainService(projectRepo, apiKeyRepo)

	// 5.2 ApiKey 领域服务
	apiKeyDomainService := apiKeyService.NewApiKeyDomainService(apiKeyRepo)

	// 5.3 Metrics 领域服务
	metricsCollectionDomainService := metricsService.NewMetricsCollectionDomainService(metricsRepo)

	// 5.4 ApiInstance 领域服务
	strategyFactory := apiInstanceStrategy.NewLoadBalancingStrategyFactory()
	affinityService := apiInstanceService.NewAffinityService()
	affinityDecorator := apiInstanceService.NewAffinityAwareStrategyDecorator(affinityService)
	apiInstanceDomainService := apiInstanceService.NewApiInstanceDomainService(apiInstanceRepo)
	apiInstanceSelectionDomainService := apiInstanceService.NewApiInstanceSelectionDomainService(apiInstanceRepo, strategyFactory, affinityDecorator)

	// 6. 初始化应用层
	authenticationAppService := appService.NewAuthenticationAppService(apiKeyDomainService)
	selectionAppService := appService.NewSelectionAppService(apiInstanceSelectionDomainService, metricsCollectionDomainService, projectDomainService)
	apiInstanceAppService := appService.NewApiInstanceAppService(apiInstanceDomainService, projectDomainService)

	// 7. 初始化数据（默认 API Key 和项目）
	initDefaultData(apiKeyDomainService, projectDomainService)

	// 8. 设置路由并启动服务
	r := router.SetupRouter(authenticationAppService, selectionAppService, apiInstanceAppService)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Info().Str("addr", addr).Msg("服务启动成功，开始监听")
	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("服务启动失败")
	}
}

// initDefaultData 初始化默认数据
func initDefaultData(apiKeyService *apiKeyService.ApiKeyDomainService, projectService *projectService.ProjectDomainService) {
	defaultApiKey := os.Getenv("DEFAULT_API_KEY")
	if defaultApiKey == "" {
		defaultApiKey = "gw_default_key_for_development"
	}

	// 检查默认 API Key 是否已存在
	if !apiKeyService.ExistApiKey(defaultApiKey) {
		if err := apiKeyService.CreateApiKey(defaultApiKey); err != nil {
			log.Error().Err(err).Msg("创建默认 API Key 失败")
		} else {
			log.Info().Str("apiKey", defaultApiKey).Msg("默认 API Key 创建成功")
		}
	} else {
		log.Info().Msg("默认 API Key 已存在，跳过创建")
	}

	// 检查默认项目是否已存在
	defaultProjectName := "default-project"
	existing, _ := projectService.GetProjectByApiKey(defaultApiKey)
	if existing == nil {
		_, err := projectService.CreateProject(defaultProjectName, "默认项目", defaultApiKey)
		if err != nil {
			log.Error().Err(err).Msg("创建默认项目失败")
		} else {
			log.Info().Str("name", defaultProjectName).Msg("默认项目创建成功")
		}
	} else {
		log.Info().Str("name", existing.Name).Msg("默认项目已存在，跳过创建")
	}
}
