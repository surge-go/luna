package types

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// ApiCreateGroupReq 创建分组请求
type ApiCreateGroupReq struct {
	ParentID *int64 `json:"parent_id"`
	Group    string `json:"group"`
	Name     string `json:"name"`
	Remark   string `json:"remark"`
}

func (r *ApiCreateGroupReq) Validate() error {
	if r == nil {
		return errors.New("request is nil")
	}

	r.Group = strings.TrimSpace(r.Group)
	r.Name = strings.TrimSpace(r.Name)
	r.Remark = strings.TrimSpace(r.Remark)

	if r.Group == "" {
		return errors.New("group is empty")
	}
	if utf8.RuneCountInString(r.Group) > 64 {
		return errors.New("group exceeds 64 characters")
	}

	if r.Name == "" {
		return errors.New("name is empty")
	}
	if utf8.RuneCountInString(r.Name) > 128 {
		return errors.New("name exceeds 128 characters")
	}
	if utf8.RuneCountInString(r.Remark) > 255 {
		return errors.New("remark exceeds 255 characters")
	}
	if r.ParentID != nil && *r.ParentID <= 0 {
		return errors.New("parent_id must be greater than zero")
	}

	return nil
}

// ApiDeleteGroupReq 删除分组请求
type ApiDeleteGroupReq struct {
}

// ApiUpdateGroupReq 更新分组请求
type ApiUpdateGroupReq struct {
}

// ApiGroupDetailResp 分组详情响应
type ApiGroupDetailResp struct {
}

// ApiCreateAPIReq 创建Api请求
type ApiCreateAPIReq struct {
}

// ApiDeleteAPIReq 删除Api请求
type ApiDeleteAPIReq struct {
}

// ApiUpdateAPIReq 更新Api请求
type ApiUpdateAPIReq struct {
}

// ApiAPIDetailResp Api详情响应
type ApiAPIDetailResp struct {
}

// ApiTreeResp Api树响应
type ApiTreeResp struct {
}
