// Package service 实现了应用层的业务编排服务。
//
// 该包属于 DDD 架构的应用层，负责编排多个领域服务的协作流程，
// 包括实例选择、调用结果上报、降级链处理等。
// 事务管理应在此层处理。
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

// SelectionAppService 选择算法应用服务。
//
// 编排 API 实例选择的完整流程，包括：
//   - 项目存在性验证
//   - 候选实例查找
//   - 指标数据获取
//   - 健康实例过滤（排除熔断实例）
//   - 负载均衡策略选择
//   - 降级链（Fallback Chain）处理
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

// SelectBestInstance 选择最佳 API 实例（支持降级）。
//
// 完整流程：
//  1. 尝试按主要标识符选择实例
//  2. 如果失败且请求中包含 FallbackChain，则依次尝试降级标识符
//  3. 如果所有降级均失败，返回 FALLBACK_EXHAUSTED 错误
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

// selectInstanceInternal 内部实例选择方法。
//
// 执行完整的实例选择流程：
//  1. 转换请求为领域命令
//  2. 验证项目存在
//  3. 查找候选实例
//  4. 获取实例指标
//  5. 过滤熔断实例
//  6. 使用策略选择最佳实例
//  7. 转换为 DTO 返回
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

// tryFallbackInstances 尝试降级实例选择。
//
// 依次遍历 FallbackChain 中的备选标识符，
// 尝试为每个标识符选择可用实例。
// 任一成功即返回，全部失败则返回 FALLBACK_EXHAUSTED 错误。
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

// ReportCallResult 上报调用结果。
//
// 将上游服务的调用结果（成功/失败、延迟、错误信息等）
// 转发给 MetricsCollectionDomainService 进行指标记录和健康状态更新。
func (s *SelectionAppService) ReportCallResult(req *request.ReportResultRequest, projectID string) error {
	log.Info().Str("instanceId", req.InstanceID).Bool("success", req.Success).Msg("应用层开始处理调用结果上报")

	cmd := assembler.ReportRequestToCommand(req, projectID)
	if err := s.metricsCollectionDomainService.RecordCallResult(cmd); err != nil {
		return err
	}

	log.Info().Msg("应用层调用结果上报处理完成")
	return nil
}
