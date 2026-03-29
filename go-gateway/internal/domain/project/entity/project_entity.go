package entity

import (
	"fmt"
	"time"
)

// ProjectStatus 项目状态
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "ACTIVE"
	ProjectStatusInactive ProjectStatus = "INACTIVE"
)

// ProjectStatusFromCode 根据代码获取项目状态
func ProjectStatusFromCode(code string) (ProjectStatus, error) {
	switch code {
	case "ACTIVE":
		return ProjectStatusActive, nil
	case "INACTIVE":
		return ProjectStatusInactive, nil
	default:
		return "", fmt.Errorf("未知的项目状态代码: %s", code)
	}
}

// ProjectEntity 项目领域实体
type ProjectEntity struct {
	ID          string        `gorm:"column:id;primaryKey" json:"id"`
	Name        string        `gorm:"column:name" json:"name"`
	Description string        `gorm:"column:description" json:"description"`
	ApiKey      string        `gorm:"column:api_key" json:"api_key"`
	Status      ProjectStatus `gorm:"column:status" json:"status"`
	CreatedAt   time.Time     `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (ProjectEntity) TableName() string {
	return "projects"
}

// NewProjectEntity 创建项目实体
func NewProjectEntity(name, description, apiKey string) *ProjectEntity {
	return &ProjectEntity{
		Name:        name,
		Description: description,
		ApiKey:      apiKey,
		Status:      ProjectStatusActive,
	}
}

// Activate 激活项目
func (p *ProjectEntity) Activate() {
	p.Status = ProjectStatusActive
}

// Deactivate 停用项目
func (p *ProjectEntity) Deactivate() {
	p.Status = ProjectStatusInactive
}

// IsActive 检查项目是否处于活跃状态
func (p *ProjectEntity) IsActive() bool {
	return p.Status == ProjectStatusActive
}

// UpdateInfo 更新项目信息
func (p *ProjectEntity) UpdateInfo(name, description string) {
	p.Name = name
	p.Description = description
}

// RegenerateApiKey 重新生成API Key
func (p *ProjectEntity) RegenerateApiKey(newApiKey string) {
	p.ApiKey = newApiKey
}
