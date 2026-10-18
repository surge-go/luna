package models

import "time"

// RoleMenu 角色菜单关联模型
type RoleMenu struct {
	ID        int64     `gorm:"primaryKey;column:id" json:"id"`
	RoleID    int64     `gorm:"column:role_id;not null;uniqueIndex:uk_sys_role_menu;index:idx_sys_role_menu_role_id" json:"role_id"`
	MenuID    int64     `gorm:"column:menu_id;not null;uniqueIndex:uk_sys_role_menu;index:idx_sys_role_menu_menu_id" json:"menu_id"`
	CreatedAt time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`

	// 关联关系
	Role *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Menu *Menu `gorm:"foreignKey:MenuID" json:"menu,omitempty"`
}

// TableName 指定表名
func (*RoleMenu) TableName() string {
	return "sys_role_menu"
}
