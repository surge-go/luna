package models

import "time"

// API API资源模型
type API struct {
	ID        int64      `gorm:"primaryKey;column:id" json:"id"`
	ParentID  *int64     `gorm:"column:parent_id;index:idx_sys_api_parent_deleted" json:"parent_id,omitempty"`
	Group     *string    `gorm:"column:group;type:varchar(64);uniqueIndex:idx_sys_api_group,where:deleted_at IS NULL AND \"group\" IS NOT NULL" json:"group,omitempty"`
	Name      string     `gorm:"column:name;type:varchar(128);not null" json:"name"`
	Method    *string    `gorm:"column:method;type:varchar(8)" json:"method,omitempty"`
	Path      *string    `gorm:"column:path;type:varchar(512)" json:"path,omitempty"`
	Sort      int        `gorm:"column:sort;not null;default:0;index:idx_sys_api_sort,where:deleted_at IS NULL" json:"sort"`
	Remark    *string    `gorm:"column:remark;type:varchar(255)" json:"remark,omitempty"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`
	DeletedAt *time.Time `gorm:"column:deleted_at;index:idx_sys_api_deleted_at" json:"deleted_at,omitempty"`

	// 关联关系
	Parent   *API    `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []*API  `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Menus    []*Menu `gorm:"many2many:sys_menu_api;foreignKey:ID;joinForeignKey:api_id;References:ID;joinReferences:menu_id" json:"menus,omitempty"`
}

// TableName 指定表名
func (*API) TableName() string {
	return "sys_api"
}

// IsGroup 判断是否是分组节点
func (a *API) IsGroup() bool {
	return a.Group != nil && *a.Group != ""
}

// IsAPI 判断是否是API节点
func (a *API) IsAPI() bool {
	return !a.IsGroup() && a.Method != nil && a.Path != nil
}

// IsRoot 判断是否是根节点
func (a *API) IsRoot() bool {
	return a.ParentID == nil
}

// HasChildren 判断是否有子节点
func (a *API) HasChildren() bool {
	return len(a.Children) > 0
}

// IsDeleted 判断是否已删除
func (a *API) IsDeleted() bool {
	return a.DeletedAt != nil
}

// GetMethodUpper 获取大写的HTTP方法
func (a *API) GetMethodUpper() string {
	if a.Method == nil {
		return ""
	}
	return *a.Method
}

// GetPath 获取路径
func (a *API) GetPath() string {
	if a.Path == nil {
		return ""
	}
	return *a.Path
}

// GetGroup 获取分组标识
func (a *API) GetGroup() string {
	if a.Group == nil {
		return ""
	}
	return *a.Group
}
