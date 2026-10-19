package types

import "time"

// CreateMenuReq 创建菜单请求
type CreateMenuReq struct {
	ParentID           *int64  `json:"parent_id" binding:"omitempty,gt=0"`                                       // 父级菜单ID
	Type               string  `json:"type" binding:"required,oneof=directory page button external"`            // 菜单类型
	PermCode           *string `json:"perm_code" binding:"omitempty,max=128"`                                   // 权限编码
	RouteName          *string `json:"route_name" binding:"omitempty,max=128"`                                  // 路由名称
	Title              string  `json:"title" binding:"required,max=128"`                                        // 菜单标题
	Path               *string `json:"path" binding:"omitempty,max=512"`                                        // 路由路径
	Component          *string `json:"component" binding:"omitempty,max=512"`                                   // 组件路径
	Link               *string `json:"link" binding:"omitempty,max=1024,url"`                                   // 外链地址
	Redirect           *string `json:"redirect" binding:"omitempty,max=512"`                                    // 重定向路径
	Icon               *string `json:"icon" binding:"omitempty,max=128"`                                        // 图标
	ActiveIcon         *string `json:"active_icon" binding:"omitempty,max=128"`                                 // 激活图标
	Sort               int     `json:"sort" binding:"omitempty,gte=0"`                                          // 排序
	Visible            *bool   `json:"visible" binding:"omitempty"`                                             // 是否可见
	HideChildrenInMenu *bool   `json:"hide_children_in_menu" binding:"omitempty"`                               // 隐藏子菜单
	KeepAlive          *bool   `json:"keep_alive" binding:"omitempty"`                                          // 是否缓存
	AffixTab           *bool   `json:"affix_tab" binding:"omitempty"`                                           // 固定标签
	Status             *int16  `json:"status" binding:"omitempty,oneof=0 1"`                                    // 状态：0-禁用 1-启用
}

// DeleteMenuReq 删除菜单请求
type DeleteMenuReq struct {
	ID      int64 `json:"id" form:"id" uri:"id" binding:"required,gt=0"` // 菜单ID
	Confirm *bool `json:"confirm" form:"confirm"`                        // 确认删除（当存在按钮时需要）
}

// UpdateMenuReq 更新菜单请求
type UpdateMenuReq struct {
	ID                 int64   `json:"id" binding:"required,gt=0"`                                              // 菜单ID
	ParentID           *int64  `json:"parent_id" binding:"omitempty,gt=0"`                                      // 父级菜单ID
	Type               *string `json:"type" binding:"omitempty,oneof=directory page button external"`          // 菜单类型
	PermCode           *string `json:"perm_code" binding:"omitempty,max=128"`                                  // 权限编码
	RouteName          *string `json:"route_name" binding:"omitempty,max=128"`                                 // 路由名称
	Title              *string `json:"title" binding:"omitempty,max=128"`                                      // 菜单标题
	Path               *string `json:"path" binding:"omitempty,max=512"`                                       // 路由路径
	Component          *string `json:"component" binding:"omitempty,max=512"`                                  // 组件路径
	Link               *string `json:"link" binding:"omitempty,max=1024,url"`                                  // 外链地址
	Redirect           *string `json:"redirect" binding:"omitempty,max=512"`                                   // 重定向路径
	Icon               *string `json:"icon" binding:"omitempty,max=128"`                                       // 图标
	ActiveIcon         *string `json:"active_icon" binding:"omitempty,max=128"`                                // 激活图标
	Sort               *int    `json:"sort" binding:"omitempty,gte=0"`                                         // 排序
	Visible            *bool   `json:"visible" binding:"omitempty"`                                            // 是否可见
	HideChildrenInMenu *bool   `json:"hide_children_in_menu" binding:"omitempty"`                              // 隐藏子菜单
	KeepAlive          *bool   `json:"keep_alive" binding:"omitempty"`                                         // 是否缓存
	AffixTab           *bool   `json:"affix_tab" binding:"omitempty"`                                          // 固定标签
	Status             *int16  `json:"status" binding:"omitempty,oneof=0 1"`                                   // 状态：0-禁用 1-启用
}

// MenuTreeResp 菜单树响应
type MenuTreeResp struct {
	ID                 int64           `json:"id"`
	ParentID           *int64          `json:"parent_id"`
	Type               string          `json:"type"`
	PermCode           *string         `json:"perm_code"`
	RouteName          *string         `json:"route_name"`
	Title              string          `json:"title"`
	Path               *string         `json:"path"`
	Component          *string         `json:"component"`
	Link               *string         `json:"link"`
	Redirect           *string         `json:"redirect"`
	Icon               *string         `json:"icon"`
	ActiveIcon         *string         `json:"active_icon"`
	Sort               int             `json:"sort"`
	Visible            bool            `json:"visible"`
	HideChildrenInMenu bool            `json:"hide_children_in_menu"`
	KeepAlive          bool            `json:"keep_alive"`
	AffixTab           bool            `json:"affix_tab"`
	Status             int16           `json:"status"`
	CreatedAt          time.Time       `json:"created_at"`
	Children           []*MenuTreeResp `json:"children"` // 子菜单
}

// MenuDetailResp 菜单详情响应
type MenuDetailResp struct {
	ID                 int64      `json:"id"`
	ParentID           *int64     `json:"parent_id"`
	Type               string     `json:"type"`
	PermCode           *string    `json:"perm_code"`
	RouteName          *string    `json:"route_name"`
	Title              string     `json:"title"`
	Path               *string    `json:"path"`
	Component          *string    `json:"component"`
	Link               *string    `json:"link"`
	Redirect           *string    `json:"redirect"`
	Icon               *string    `json:"icon"`
	ActiveIcon         *string    `json:"active_icon"`
	Sort               int        `json:"sort"`
	Visible            bool       `json:"visible"`
	HideChildrenInMenu bool       `json:"hide_children_in_menu"`
	KeepAlive          bool       `json:"keep_alive"`
	AffixTab           bool       `json:"affix_tab"`
	Status             int16      `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          *time.Time `json:"updated_at"`
}

// CreateButtonReq 创建按钮请求
type CreateButtonReq struct {
	ParentID *int64  `json:"parent_id" binding:"required,gt=0"`         // 父级菜单ID（按钮必须属于某个菜单）
	PermCode string  `json:"perm_code" binding:"required,max=128"`      // 权限编码
	Title    string  `json:"title" binding:"required,max=128"`          // 按钮标题
	Sort     int     `json:"sort" binding:"omitempty,gte=0"`            // 排序
	Status   *int16  `json:"status" binding:"omitempty,oneof=0 1"`      // 状态：0-禁用 1-启用
}

// DeleteButtonReq 删除按钮请求
type DeleteButtonReq struct {
	ID int64 `json:"id" form:"id" uri:"id" binding:"required,gt=0"` // 按钮ID
}

// UpdateButtonReq 更新按钮请求
type UpdateButtonReq struct {
	ID       int64   `json:"id" binding:"required,gt=0"`               // 按钮ID
	ParentID *int64  `json:"parent_id" binding:"omitempty,gt=0"`       // 父级菜单ID
	PermCode *string `json:"perm_code" binding:"omitempty,max=128"`    // 权限编码
	Title    *string `json:"title" binding:"omitempty,max=128"`        // 按钮标题
	Sort     *int    `json:"sort" binding:"omitempty,gte=0"`           // 排序
	Status   *int16  `json:"status" binding:"omitempty,oneof=0 1"`     // 状态：0-禁用 1-启用
}

// ButtonListResp 按钮列表响应
type ButtonListResp struct {
	Items []*ButtonItem `json:"items"` // 按钮列表
}

// ButtonItem 按钮项
type ButtonItem struct {
	ID        int64      `json:"id"`
	ParentID  *int64     `json:"parent_id"`
	PermCode  string     `json:"perm_code"`
	Title     string     `json:"title"`
	Sort      int        `json:"sort"`
	Status    int16      `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// ButtonDetailResp 按钮详情响应
type ButtonDetailResp struct {
	ID        int64      `json:"id"`
	ParentID  *int64     `json:"parent_id"`
	Type      string     `json:"type"` // 固定为 "button"
	PermCode  string     `json:"perm_code"`
	Title     string     `json:"title"`
	Sort      int        `json:"sort"`
	Status    int16      `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

// MenuOptionsResp 菜单选项响应
type MenuOptionsResp struct {
	ID       int64              `json:"id"`
	ParentID *int64             `json:"parent_id"`
	Title    string             `json:"title"`
	Type     string             `json:"type"`
	Children []*MenuOptionsResp `json:"children"` // 子菜单选项
}
