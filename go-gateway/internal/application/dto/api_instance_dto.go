package dto

import (
	"time"

	metricsEntity "github.com/lucky-aeon/api-premium-gateway/go-gateway/internal/domain/metrics/entity"
)

// ApiInstanceDTO API实例数据传输对象
type ApiInstanceDTO struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"projectId"`
	ProjectName   string                 `json:"projectName,omitempty"`
	UserID        string                 `json:"userId,omitempty"`
	ApiIdentifier string                 `json:"apiIdentifier"`
	ApiType       string                 `json:"apiType"`
	BusinessID    string                 `json:"businessId"`
	RoutingParams metricsEntity.JSONBMap `json:"routingParams,omitempty"`
	Status        string                 `json:"status"`
	Metadata      metricsEntity.JSONBMap `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
}
