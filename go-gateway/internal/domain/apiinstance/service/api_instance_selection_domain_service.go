// Package service 实现了 API Instance 领域的核心业务逻辑。
//
// 该包属于 DDD 架构的领域层，包含：
//   - ApiInstanceSelectionDomainService：实例选择算法编排
//   - ApiInstanceDomainService：实例 CRUD 管理
//   - AffinityService：亲和性绑定缓存管理
//   - AffinityAwareStrategyDecorator：亲和性装饰器
package service

import (
	"github.com/rs/zerolog/log"

	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/command"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/repository"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/strategy"
	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/infrastructure/exception"
)

// ApiInstanceSelectionDomainService API 实例选择领域服务。
//
// 负责编排实例选择的核心流程：候选实例查找、健康实例过滤、
// 负载均衡策略选择（通过策略工厂）和亲和性装饰（通过装饰器）。
type ApiInstanceSelectionDomainService struct {
	apiInstanceRepository repository.ApiInstanceRepository
	strategyFactory       *strategy.LoadBalancingStrategyFactory
	affinityDecorator     *AffinityAwareStrategyDecorator
}

// NewApiInstanceSelectionDomainService 创建API实例选择领域服务
func NewApiInstanceSelectionDomainService(
	repo repository.ApiInstanceRepository,
	factory *strategy.LoadBalancingStrategyFactory,
	decorator *AffinityAwareStrategyDecorator,
) *ApiInstanceSelectionDomainService {
	return &ApiInstanceSelectionDomainService{
		apiInstanceRepository: repo,
		strategyFactory:       factory,
		affinityDecorator:     decorator,
	}
}

// FindCandidateInstances 查找候选实例。
//
// 根据项目 ID、API 类型、API 标识符和用户 ID 查询状态为 ACTIVE 的实例。
func (s *ApiInstanceSelectionDomainService) FindCandidateInstances(cmd *command.InstanceSelectionCommand) ([]*entity.ApiInstanceEntity, error) {
	candidates, err := s.apiInstanceRepository.SelectCandidates(
		cmd.ProjectID,
		entity.ApiType(cmd.ApiType),
		cmd.ApiIdentifier,
		cmd.UserID,
	)
	if err != nil {
		return nil, err
	}

	log.Debug().Str("projectId", cmd.ProjectID).Str("apiIdentifier", cmd.ApiIdentifier).Str("apiType", cmd.ApiType).Int("count", len(candidates)).Msg("查找候选实例")
	return candidates, nil
}

// FilterHealthyInstances 过滤掉被熔断的实例。
//
// 遍历候选实例，排除指标状态为 CIRCUIT_BREAKER_OPEN 的实例。
// 对于没有指标数据的新实例，默认保留。
func (s *ApiInstanceSelectionDomainService) FilterHealthyInstances(candidates []*entity.ApiInstanceEntity, metricsMap map[string]*metricsEntity.InstanceMetricsEntity) []*entity.ApiInstanceEntity {
	var result []*entity.ApiInstanceEntity
	for _, instance := range candidates {
		metrics, ok := metricsMap[instance.ID]
		if ok && metrics.IsCircuitBreakerOpen() {
			log.Debug().Str("instanceId", instance.ID).Str("businessId", instance.BusinessID).Msg("实例被熔断，过滤掉")
			continue
		}
		result = append(result, instance)
	}
	return result
}

// SelectInstanceWithStrategy 使用策略选择最佳实例。
//
// 通过策略工厂获取负载均衡策略，然后通过亲和性装饰器执行实例选择。
// 如果请求包含亲和性要求，装饰器会在策略选择之上叠加绑定逻辑。
func (s *ApiInstanceSelectionDomainService) SelectInstanceWithStrategy(
	healthyInstances []*entity.ApiInstanceEntity,
	metricsMap map[string]*metricsEntity.InstanceMetricsEntity,
	cmd *command.InstanceSelectionCommand,
) (*entity.ApiInstanceEntity, error) {
	log.Info().Int("count", len(healthyInstances)).Str("strategy", string(cmd.LoadBalancingType)).Msg("开始使用策略选择最佳API实例")

	if len(healthyInstances) == 0 {
		return nil, exception.NewBusinessErrorWithCode("NO_HEALTHY_INSTANCE", "没有健康的API实例可供选择")
	}

	// 使用亲和性感知的策略选择实例
	strat, err := s.strategyFactory.GetStrategy(entity.LoadBalancingTypeRoundRobin)
	if err != nil {
		return nil, err
	}

	selected, err := s.affinityDecorator.SelectInstanceWithAffinity(
		healthyInstances,
		metricsMap,
		strat,
		cmd.AffinityContext,
	)
	if err != nil {
		return nil, err
	}

	if cmd.HasAffinityRequirement() {
		log.Info().Str("businessId", selected.BusinessID).Str("instanceId", selected.ID).Str("strategy", string(cmd.LoadBalancingType)).Str("affinity", cmd.AffinityContext.GetBindingKey()).Msg("选择API实例成功（含亲和性）")
	} else {
		log.Info().Str("businessId", selected.BusinessID).Str("instanceId", selected.ID).Str("strategy", string(cmd.LoadBalancingType)).Msg("选择API实例成功")
	}

	return selected, nil
}
