package command

import (
	"github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/apiinstance/entity"
)

// InstanceSelectionCommand 实例选择命令
type InstanceSelectionCommand struct {
	ProjectID         string                   `json:"project_id"`
	UserID            string                   `json:"user_id"`
	ApiIdentifier     string                   `json:"api_identifier"`
	ApiType           string                   `json:"api_type"`
	LoadBalancingType entity.LoadBalancingType  `json:"load_balancing_type"`
	AffinityContext   *entity.AffinityContext   `json:"affinity_context"`
}

// NewInstanceSelectionCommand 创建实例选择命令
func NewInstanceSelectionCommand(projectID, userID, apiIdentifier, apiType string) *InstanceSelectionCommand {
	return &InstanceSelectionCommand{
		ProjectID:         projectID,
		UserID:            userID,
		ApiIdentifier:     apiIdentifier,
		ApiType:           apiType,
		LoadBalancingType: entity.LoadBalancingTypeSmart,
	}
}

// HasAffinityRequirement 检查是否有亲和性要求
func (c *InstanceSelectionCommand) HasAffinityRequirement() bool {
	return c.AffinityContext != nil && c.AffinityContext.IsValid()
}
