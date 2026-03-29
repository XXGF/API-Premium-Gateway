package service

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/assembler"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/application/dto"
	apiInstanceService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/service"
	metricsService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/service"
	projectService "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/project/service"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/infrastructure/exception"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/interfaces/request"
)

// SelectionAppService 选择算法应用服务
type SelectionAppService struct {
	apiInstanceSelectionDomainService *apiInstanceService.ApiInstanceSelectionDomainService
	metricsCollectionDomainService   *metricsService.MetricsCollectionDomainService
	projectDomainService             *projectService.ProjectDomainService
}

// NewSelectionAppService 创建选择算法应用服务
func NewSelectionAppService(
	selectionService *apiInstanceService.ApiInstanceSelectionDomainService,
	metricsService *metricsService.MetricsCollectionDomainService,
	projectService *projectService.ProjectDomainService,
) *SelectionAppService {
	return &SelectionAppService{
		apiInstanceSelectionDomainService: selectionService,
		metricsCollectionDomainService:   metricsService,
		projectDomainService:             projectService,
	}
}

// SelectBestInstance 选择最佳API实例（支持降级）
func (s *SelectionAppService) SelectBestInstance(req *request.SelectInstanceRequest, currentProjectID string) (*dto.ApiInstanceDTO, error) {
	log.Info().Str("apiIdentifier", req.ApiIdentifier).Msg("应用层开始选择API实例")

	result, err := s.selectInstanceInternal(req, currentProjectID)
	if err != nil {
		// 如果正常选择失败且有降级链，则尝试降级
		if req.HasFallbackChain() {
			if bizErr, ok := err.(*exception.BusinessError); ok {
				if bizErr.ErrorCode == "NO_AVAILABLE_INSTANCE" || bizErr.ErrorCode == "NO_HEALTHY_INSTANCE" {
					log.Warn().Str("error", err.Error()).Msg("主要实例选择失败，开始尝试降级")
					return s.tryFallbackInstances(req, currentProjectID)
				}
			}
		}
		return nil, err
	}
	return result, nil
}

// selectInstanceInternal 内部实例选择方法
func (s *SelectionAppService) selectInstanceInternal(req *request.SelectInstanceRequest, currentProjectID string) (*dto.ApiInstanceDTO, error) {
	// 1. 转换为领域命令
	cmd := assembler.SelectRequestToCommand(req, currentProjectID)

	// 2. 验证项目存在
	if err := s.projectDomainService.ValidateProjectExists(cmd.ProjectID); err != nil {
		return nil, err
	}

	// 3. 查找候选实例
	candidates, err := s.apiInstanceSelectionDomainService.FindCandidateInstances(cmd)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, exception.NewBusinessErrorWithCode("NO_AVAILABLE_INSTANCE",
			fmt.Sprintf("没有可用的API实例: projectId=%s, apiIdentifier=%s, apiType=%s",
				cmd.ProjectID, cmd.ApiIdentifier, cmd.ApiType))
	}

	// 4. 获取实例指标
	instanceIDs := make([]string, len(candidates))
	for i, c := range candidates {
		instanceIDs[i] = c.ID
	}
	metricsMap, err := s.metricsCollectionDomainService.GetInstanceMetrics(instanceIDs)
	if err != nil {
		return nil, err
	}

	// 5. 过滤掉被熔断的实例
	healthyInstances := s.apiInstanceSelectionDomainService.FilterHealthyInstances(candidates, metricsMap)
	if len(healthyInstances) == 0 {
		return nil, exception.NewBusinessErrorWithCode("NO_HEALTHY_INSTANCE", "所有API实例都不可用或被熔断")
	}

	// 6. 使用策略选择最佳实例
	selectedEntity, err := s.apiInstanceSelectionDomainService.SelectInstanceWithStrategy(healthyInstances, metricsMap, cmd)
	if err != nil {
		return nil, err
	}

	// 7. 转换为DTO返回
	result := assembler.ApiInstanceToDTO(selectedEntity)
	log.Info().Str("businessId", result.BusinessID).Str("instanceId", result.ID).Msg("应用层选择API实例成功")
	return result, nil
}

// tryFallbackInstances 尝试降级实例选择
func (s *SelectionAppService) tryFallbackInstances(req *request.SelectInstanceRequest, currentProjectID string) (*dto.ApiInstanceDTO, error) {
	log.Info().Strs("fallbackChain", req.FallbackChain).Msg("开始尝试降级实例选择")

	for i, fallbackBusinessID := range req.FallbackChain {
		fallbackReq := &request.SelectInstanceRequest{
			UserID:        req.UserID,
			ApiIdentifier: fallbackBusinessID,
			ApiType:       req.ApiType,
			AffinityKey:   req.AffinityKey,
			AffinityType:  req.AffinityType,
		}

		log.Info().Int("index", i+1).Str("businessId", fallbackBusinessID).Msg("尝试降级到实例")

		result, err := s.selectInstanceInternal(fallbackReq, currentProjectID)
		if err == nil {
			log.Info().Str("businessId", result.BusinessID).Str("instanceId", result.ID).Msg("降级成功")
			return result, nil
		}

		log.Warn().Int("index", i+1).Str("businessId", fallbackBusinessID).Str("error", err.Error()).Msg("降级失败")

		if i >= len(req.FallbackChain)-1 {
			return nil, exception.NewBusinessErrorWithCode("FALLBACK_EXHAUSTED",
				fmt.Sprintf("所有降级实例都不可用，主实例和%d个降级实例均失败", len(req.FallbackChain)))
		}
	}

	return nil, exception.NewBusinessErrorWithCode("FALLBACK_EXHAUSTED", "降级链为空或处理异常")
}

// ReportCallResult 上报调用结果
func (s *SelectionAppService) ReportCallResult(req *request.ReportResultRequest, projectID string) error {
	log.Info().Str("instanceId", req.InstanceID).Bool("success", req.Success).Msg("应用层开始处理调用结果上报")

	cmd := assembler.ReportRequestToCommand(req, projectID)
	if err := s.metricsCollectionDomainService.RecordCallResult(cmd); err != nil {
		return err
	}

	log.Info().Msg("应用层调用结果上报处理完成")
	return nil
}
