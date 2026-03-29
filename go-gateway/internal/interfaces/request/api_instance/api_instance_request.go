package api_instance

// ApiInstanceCreateRequest 创建API实例请求
type ApiInstanceCreateRequest struct {
	UserID        string                 `json:"userId"`
	ApiIdentifier string                 `json:"apiIdentifier" binding:"required"`
	ApiType       string                 `json:"apiType" binding:"required"`
	BusinessID    string                 `json:"businessId" binding:"required"`
	RoutingParams map[string]interface{} `json:"routingParams"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ApiInstanceUpdateRequest 更新API实例请求
type ApiInstanceUpdateRequest struct {
	UserID        string                 `json:"userId"`
	ApiIdentifier string                 `json:"apiIdentifier"`
	RoutingParams map[string]interface{} `json:"routingParams"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ApiInstanceBatchCreateRequest 批量创建API实例请求
type ApiInstanceBatchCreateRequest struct {
	Instances []*ApiInstanceCreateRequest `json:"instances" binding:"required,dive"`
}

// ApiInstanceBatchDeleteRequest 批量删除API实例请求
type ApiInstanceBatchDeleteRequest struct {
	Instances []*ApiInstanceDeleteItem `json:"instances" binding:"required,dive"`
}

// ApiInstanceDeleteItem 删除项
type ApiInstanceDeleteItem struct {
	ApiType    string `json:"apiType" binding:"required"`
	BusinessID string `json:"businessId" binding:"required"`
}
