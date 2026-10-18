package handler

import (
	"context"

	"github.com/surge-go/aurora"
	"github.com/surge-go/luna/internal/modules/system/types"
	"github.com/surge-go/luna/pkg/dto"
)

type ApiService interface {
	// CreateGroup 创建分组
	CreateGroup(ctx context.Context, req *types.ApiCreateGroupReq) error
	// DeleteGroup 删除分组
	DeleteGroup(ctx context.Context, req *dto.IDReq) error
	// UpdateGroup 更新分组
	UpdateGroup(ctx context.Context, req *types.ApiUpdateGroupReq) error
	// GroupDetail 获取分组详情
	GroupDetail(ctx context.Context, req *dto.IDReq) (*types.ApiGroupDetailResp, error)

	// CreateApi 创建Api
	CreateApi(ctx context.Context, req *types.ApiCreateAPIReq) error
	// DeleteApi 删除Api
	DeleteApi(ctx context.Context, req *types.ApiDeleteAPIReq) error
	// UpdateApi 更新Api
	UpdateApi(ctx context.Context, req *types.ApiUpdateAPIReq) error
	// APIDetail 获取Api详情
	APIDetail(ctx context.Context, req *dto.IDReq) (*types.ApiAPIDetailResp, error)

	// Tree 获取Api树
	Tree(ctx context.Context) ([]*types.ApiTreeResp, error)
	// Options 获取Api选项
	Options(ctx context.Context) ([]*types.ApiOptionsResp, error)
}

type ApiHandler struct {
	svc ApiService
}

func NewApiHandler(svc ApiService) *ApiHandler {
	return &ApiHandler{svc: svc}
}

func (h *ApiHandler) CreateGroup(ctx *aurora.Context, req *types.ApiCreateGroupReq) error {
	return h.svc.CreateGroup(ctx.Context(), req)
}

func (h *ApiHandler) DeleteGroup(ctx *aurora.Context, req *dto.IDReq) error {
	return h.svc.DeleteGroup(ctx.Context(), req)
}

func (h *ApiHandler) UpdateGroup(ctx *aurora.Context, req *types.ApiUpdateGroupReq) error {
	return h.svc.UpdateGroup(ctx.Context(), req)
}

func (h *ApiHandler) GroupDetail(ctx *aurora.Context, req *dto.IDReq) (*types.ApiGroupDetailResp, error) {
	return h.svc.GroupDetail(ctx.Context(), req)
}

func (h *ApiHandler) CreateApi(ctx *aurora.Context, req *types.ApiCreateAPIReq) error {
	return h.svc.CreateApi(ctx.Context(), req)
}

func (h *ApiHandler) DeleteApi(ctx *aurora.Context, req *types.ApiDeleteAPIReq) error {
	return h.svc.DeleteApi(ctx.Context(), req)
}

func (h *ApiHandler) UpdateApi(ctx *aurora.Context, req *types.ApiUpdateAPIReq) error {
	return h.svc.UpdateApi(ctx.Context(), req)
}

func (h *ApiHandler) APIDetail(ctx *aurora.Context, req *dto.IDReq) (*types.ApiAPIDetailResp, error) {
	return h.svc.APIDetail(ctx.Context(), req)
}

func (h *ApiHandler) Tree(ctx *aurora.Context) {
	if list, err := h.svc.Tree(ctx.Context()); err != nil {
		ctx.Fail(err)
	} else {
		ctx.OK(list)
	}
}

func (h *ApiHandler) Options(ctx *aurora.Context) {
	if list, err := h.svc.Options(ctx.Context()); err != nil {
		ctx.Fail(err)
	} else {
		ctx.OK(list)
	}
}
