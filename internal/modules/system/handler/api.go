package handler

import (
	"context"

	"github.com/surge-go/luna/internal/modules/system/types"
	"github.com/surge-go/luna/pkg/dto"
)

type ApiService interface {
	// CreateGroup 创建分组
	CreateGroup(ctx context.Context, req *types.ApiCreateGroupReq) error
	// DeleteGroup 删除分组
	DeleteGroup(ctx context.Context, req *types.ApiDeleteGroupReq) error
	// UpdateGroup 更新分组
	UpdateGroup(ctx context.Context, req *types.ApiUpdateGroupReq) error
	// GroupDetail 获取分组详情
	GroupDetail(ctx context.Context, req *dto.IDReq) (*types.ApiGroupDetailResp, error)

	// CreateAPI 创建Api
	CreateApi(ctx context.Context, req *types.ApiCreateAPIReq) error
	// DeleteAPI 删除Api
	DeleteApi(ctx context.Context, req *types.ApiDeleteAPIReq) error
	// UpdateAPI 更新Api
	UpdateApi(ctx context.Context, req *types.ApiUpdateAPIReq) error
	// APIetail 获取Api详情
	APIDetail(ctx context.Context, req *dto.IDReq) (*types.ApiAPIDetailResp, error)

	// Tree 获取Api树
	Tree(ctx context.Context) (*types.ApiTreeResp, error)
}

type ApiHandler struct {
	svc ApiService
}

func NewApiHandler(svc ApiService) *ApiHandler {
	return &ApiHandler{svc: svc}
}
