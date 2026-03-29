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

// ApiInstanceSelectionDomainService API实例选择领域服务
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

// FindCandidateInstances 查找候选实例
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

// FilterHealthyInstances 过滤掉被熔断的实例
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

// SelectInstanceWithStrategy 使用策略选择最佳实例
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
