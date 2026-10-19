package handler

import (
	"context"

	"github.com/surge-go/aurora"
	"github.com/surge-go/luna/internal/modules/system/types"
	"github.com/surge-go/luna/pkg/dto"
)

// MenuService 菜单服务
type MenuService interface {
	// CreateMenu 创建菜单
	CreateMenu(ctx context.Context, req *types.CreateMenuReq) error
	// DeleteMenu 删除菜单
	DeleteMenu(ctx context.Context, req *types.DeleteMenuReq) error
	// UpdateMenu 更新菜单
	UpdateMenu(ctx context.Context, req *types.UpdateMenuReq) error
	// MenuTree 获取菜单树
	MenuTree(ctx context.Context) ([]*types.MenuTreeResp, error)
	// MenuDetail 获取菜单详情
	MenuDetail(ctx context.Context, req *dto.IDReq) (*types.MenuDetailResp, error)

	// CreateButton 创建按钮
	CreateButton(ctx context.Context, req *types.CreateButtonReq) error
	// DeleteButton 删除按钮
	DeleteButton(ctx context.Context, req *types.DeleteButtonReq) error
	// UpdateButton 更新按钮
	UpdateButton(ctx context.Context, req *types.UpdateButtonReq) error
	// ButtonList 获取按钮列表
	ButtonList(ctx context.Context) (*types.ButtonListResp, error)
	// ButtonDetail 获取按钮详情
	ButtonDetail(ctx context.Context, req *dto.IDReq) (*types.ButtonDetailResp, error)

	// Options 获取菜单选项
	Options(ctx context.Context) ([]*types.MenuOptionsResp, error)
}

// MenuHandler 菜单处理
type MenuHandler struct {
	svc MenuService
}

func NewMenuHandler(svc MenuService) *MenuHandler {
	return &MenuHandler{svc: svc}
}

// CreateMenu 创建菜单
func (h *MenuHandler) CreateMenu(ctx *aurora.Context, req *types.CreateMenuReq) error {
	return h.svc.CreateMenu(ctx.Context(), req)
}

// DeleteMenu 删除菜单
func (h *MenuHandler) DeleteMenu(ctx *aurora.Context, req *types.DeleteMenuReq) error {
	return h.svc.DeleteMenu(ctx.Context(), req)
}

// UpdateMenu 更新菜单
func (h *MenuHandler) UpdateMenu(ctx *aurora.Context, req *types.UpdateMenuReq) error {
	return h.svc.UpdateMenu(ctx.Context(), req)
}

// MenuTree 获取菜单树
func (h *MenuHandler) MenuTree(ctx *aurora.Context) {
	if tree, err := h.svc.MenuTree(ctx.Context()); err != nil {
		ctx.Fail(err)
	} else {
		ctx.OK(tree)
	}
}

// MenuDetail 获取菜单详情
func (h *MenuHandler) MenuDetail(ctx *aurora.Context, req *dto.IDReq) (*types.MenuDetailResp, error) {
	return h.svc.MenuDetail(ctx.Context(), req)
}

// CreateButton 创建按钮
func (h *MenuHandler) CreateButton(ctx *aurora.Context, req *types.CreateButtonReq) error {
	return h.svc.CreateButton(ctx.Context(), req)
}

// DeleteButton 删除按钮
func (h *MenuHandler) DeleteButton(ctx *aurora.Context, req *types.DeleteButtonReq) error {
	return h.svc.DeleteButton(ctx.Context(), req)
}

// UpdateButton 更新按钮
func (h *MenuHandler) UpdateButton(ctx *aurora.Context, req *types.UpdateButtonReq) error {
	return h.svc.UpdateButton(ctx.Context(), req)
}

// ButtonList 获取按钮列表
func (h *MenuHandler) ButtonList(ctx *aurora.Context) {
	if list, err := h.svc.ButtonList(ctx.Context()); err != nil {
		ctx.Fail(err)
	} else {
		ctx.OK(list)
	}
}

// ButtonDetail 获取按钮详情
func (h *MenuHandler) ButtonDetail(ctx *aurora.Context, req *dto.IDReq) (*types.ButtonDetailResp, error) {
	return h.svc.ButtonDetail(ctx.Context(), req)
}

// Options 获取菜单选项
func (h *MenuHandler) Options(ctx *aurora.Context) {
	if options, err := h.svc.Options(ctx.Context()); err != nil {
		ctx.Fail(err)
	} else {
		ctx.OK(options)
	}
}
