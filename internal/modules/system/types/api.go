package types

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/surge-go/luna/pkg/dto"
)

// ApiCreateGroupReq 创建分组请求
type ApiCreateGroupReq struct {
	ParentID *int64 `json:"parent_id"`
	Group    string `json:"group"`
	Name     string `json:"name"`
	Sort     int    `json:"sort"`
	Remark   string `json:"remark"`
}

func (r *ApiCreateGroupReq) Validate() error {
	if r == nil {
		return errors.New("请求参数为空")
	}

	r.Group = strings.TrimSpace(r.Group)
	r.Name = strings.TrimSpace(r.Name)
	r.Remark = strings.TrimSpace(r.Remark)

	if r.Group == "" {
		return errors.New("分组标识不能为空")
	}
	if utf8.RuneCountInString(r.Group) > 64 {
		return errors.New("分组标识长度不能超过64个字符")
	}

	if r.Name == "" {
		return errors.New("分组名称不能为空")
	}
	if utf8.RuneCountInString(r.Name) > 128 {
		return errors.New("分组名称长度不能超过128个字符")
	}
	if utf8.RuneCountInString(r.Remark) > 255 {
		return errors.New("备注长度不能超过255个字符")
	}
	if r.ParentID != nil && *r.ParentID <= 0 {
		return errors.New("父级分组ID必须大于0")
	}

	return nil
}

// ApiUpdateGroupReq 更新分组请求
type ApiUpdateGroupReq struct {
	dto.IDReq
	ApiCreateGroupReq
}

// ApiGroupDetailResp 分组详情响应
type ApiGroupDetailResp struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parent_id"`
	Group     string    `json:"group"`
	Name      string    `json:"name"`
	Sort      int       `json:"sort"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

// ApiCreateAPIReq 创建Api请求
type ApiCreateAPIReq struct {
	ParentID *int64 `json:"parent_id"`
	Path     string `json:"path"`
	Method   string `json:"method"`
	Name     string `json:"name"`
	Sort     int    `json:"sort"`
	Remark   string `json:"remark"`
}

func (r *ApiCreateAPIReq) Validate() error {
	if r == nil {
		return errors.New("请求参数为空")
	}

	r.Path = strings.TrimSpace(r.Path)
	r.Method = strings.TrimSpace(r.Method)
	r.Name = strings.TrimSpace(r.Name)
	r.Remark = strings.TrimSpace(r.Remark)

	if r.ParentID != nil && *r.ParentID <= 0 {
		return errors.New("父级分组ID必须大于0")
	}

	if r.Path == "" {
		return errors.New("路径不能为空")
	}
	if utf8.RuneCountInString(r.Path) > 256 {
		return errors.New("路径长度不能超过256个字符")
	}

	switch r.Method {
	case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS":
		break
	default:
		return errors.New("方法必须是GET、POST、PUT、DELETE、PATCH、HEAD或OPTIONS")
	}
	if utf8.RuneCountInString(r.Method) > 64 {
		return errors.New("方法长度不能超过64个字符")
	}
	if r.Name == "" {
		return errors.New("名称不能为空")
	}
	if utf8.RuneCountInString(r.Name) > 128 {
		return errors.New("名称长度不能超过128个字符")
	}
	if utf8.RuneCountInString(r.Remark) > 255 {
		return errors.New("备注长度不能超过255个字符")
	}
	return nil
}

// ApiDeleteAPIReq 删除Api请求
type ApiDeleteAPIReq struct {
	dto.IDReq
	Confirm bool `json:"confirm"`
}

// ApiUpdateAPIReq 更新Api请求
type ApiUpdateAPIReq struct {
	dto.IDReq
	ApiCreateAPIReq
}

// ApiAPIDetailResp Api详情响应
type ApiAPIDetailResp struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parent_id"`
	Name      string    `json:"name"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Sort      int       `json:"sort"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

// ApiTreeResp API树响应
type ApiTreeResp struct {
	ID        int64          `json:"id"`
	ParentID  *int64         `json:"parent_id"`
	Group     *string        `json:"group"`
	Name      string         `json:"name"`
	Sort      int            `json:"sort"`
	Remark    *string        `json:"remark"`
	CreatedAt time.Time      `json:"created_at"`
	Children  []*ApiTreeResp `json:"children"`
	Apis      []*ApiTreeNode `json:"apis"`
}

// ApiTreeNode API树节点
type ApiTreeNode struct {
	ID        int64     `json:"id"`
	ParentID  int64     `json:"parent_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "group" 或 "api"
	Sort      int       `json:"sort"`
	Method    *string   `json:"method"`
	Path      *string   `json:"path"`
	Remark    *string   `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
}

type ApiOptionsResp struct {
	ID       int64             `json:"id"`
	ParentID *int64            `json:"parent_id"`
	Group    *string           `json:"group"`
	Name     string            `json:"name"`
	Children []*ApiOptionsResp `json:"children"`
	Apis     []*ApiOptionsNode `json:"apis"`
}

type ApiOptionsNode struct {
	ID       int64  `json:"id"`
	ParentID int64  `json:"parent_id"`
	Name     string `json:"name"`
}
