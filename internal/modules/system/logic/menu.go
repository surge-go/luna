package logic

import (
	"context"
	"errors"

	"github.com/surge-go/aurora"
	"github.com/surge-go/aurora/core/logger"
	"github.com/surge-go/luna/internal/models"
	"github.com/surge-go/luna/internal/modules/system/svc"
	"github.com/surge-go/luna/internal/modules/system/types"
	"github.com/surge-go/luna/pkg/dto"
	"github.com/surge-go/luna/pkg/ptr"
	"gorm.io/gorm"
)

type MenuLogic struct {
	*svc.ServiceContext
}

func (s *MenuLogic) CreateMenu(ctx context.Context, req *types.CreateMenuReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 验证父级菜单是否存在
		if req.ParentID != nil {
			var parent models.Menu
			if err := tx.Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					s.Log.WarnContext(ctx, "menu.CreateMenu:父级菜单不存在", logger.Int64("parent_id", *req.ParentID))
					return aurora.NewError(400, "父级菜单不存在")
				}
				s.Log.ErrorContext(ctx, "menu.CreateMenu:查询父级菜单失败", logger.Error(err))
				return aurora.ErrInternal
			}

			// 按钮不能作为父级菜单
			if parent.IsButton() {
				s.Log.WarnContext(ctx, "menu.CreateMenu:按钮不能作为父级菜单",
					logger.Int64("parent_id", *req.ParentID),
					logger.String("parent_type", string(parent.Type)))
				return aurora.NewError(400, "按钮不能作为父级菜单")
			}
		}

		// 2. 验证权限编码唯一性
		if req.PermCode != nil && *req.PermCode != "" {
			var count int64
			if err := tx.Model(&models.Menu{}).
				Where("perm_code = ?", *req.PermCode).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "menu.CreateMenu:查询权限编码失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				s.Log.WarnContext(ctx, "menu.CreateMenu:权限编码已存在", logger.String("perm_code", *req.PermCode))
				return aurora.NewError(400, "权限编码已存在")
			}
		}

		// 3. 验证路由名称唯一性
		if req.RouteName != nil && *req.RouteName != "" {
			var count int64
			if err := tx.Model(&models.Menu{}).
				Where("route_name = ?", *req.RouteName).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "menu.CreateMenu:查询路由名称失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				s.Log.WarnContext(ctx, "menu.CreateMenu:路由名称已存在", logger.String("route_name", *req.RouteName))
				return aurora.NewError(400, "路由名称已存在")
			}
		}

		// 4. 根据类型验证必填字段
		if err := s.validateMenuTypeFields(req); err != nil {
			s.Log.WarnContext(ctx, "menu.CreateMenu:菜单类型字段验证失败", logger.Error(err))
			return aurora.ErrBadRequest
		}

		// 5. 设置默认值并创建菜单
		menu := &models.Menu{
			ParentID:           req.ParentID,
			Type:               models.MenuType(req.Type),
			PermCode:           req.PermCode,
			RouteName:          req.RouteName,
			Title:              req.Title,
			Path:               req.Path,
			Component:          req.Component,
			Link:               req.Link,
			Redirect:           req.Redirect,
			Icon:               req.Icon,
			ActiveIcon:         req.ActiveIcon,
			Sort:               req.Sort,
			Visible:            ptr.BoolValue(req.Visible, true),
			HideChildrenInMenu: ptr.BoolValue(req.HideChildrenInMenu, false),
			KeepAlive:          ptr.BoolValue(req.KeepAlive, false),
			AffixTab:           ptr.BoolValue(req.AffixTab, false),
			Status:             ptr.Int16Value(req.Status, 1),
		}

		if err := tx.Create(menu).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.CreateMenu:创建菜单失败", logger.Error(err))
			return aurora.ErrInternal
		}

		s.Log.InfoContext(ctx, "menu.CreateMenu:创建菜单成功",
			logger.Int64("menu_id", menu.ID),
			logger.String("title", menu.Title),
			logger.String("type", string(menu.Type)))

		return nil
	})
}

func (s *MenuLogic) DeleteMenu(ctx context.Context, req *types.DeleteMenuReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 检查菜单是否存在
		var menu models.Menu
		if err := tx.Where("id = ?", req.ID).First(&menu).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.Log.WarnContext(ctx, "menu.DeleteMenu:菜单不存在", logger.Int64("menu_id", req.ID))
				return aurora.NewError(404, "菜单不存在")
			}
			s.Log.ErrorContext(ctx, "menu.DeleteMenu:查询菜单失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 2. 如果是目录，检查是否有子菜单（包括页面和按钮）
		if menu.Type == models.MenuTypeDirectory {
			var childCount int64
			if err := tx.Model(&models.Menu{}).Where("parent_id = ?", req.ID).Count(&childCount).Error; err != nil {
				s.Log.ErrorContext(ctx, "menu.DeleteMenu:查询子菜单失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if childCount > 0 {
				s.Log.WarnContext(ctx, "menu.DeleteMenu:目录下存在子菜单，请先删除子菜单",
					logger.Int64("menu_id", req.ID),
					logger.Int64("child_count", childCount))
				return aurora.NewError(400, "目录下存在子菜单，请先删除子菜单")
			}
		}

		// 3. 如果是页面，检查是否有按钮
		if menu.Type == models.MenuTypePage {
			var buttonCount int64
			if err := tx.Model(&models.Menu{}).
				Where("parent_id = ? AND type = ?", req.ID, models.MenuTypeButton).
				Count(&buttonCount).Error; err != nil {
				s.Log.ErrorContext(ctx, "menu.DeleteMenu:查询按钮失败", logger.Error(err))
				return aurora.ErrInternal
			}

			if buttonCount > 0 {
				// 需要用户确认
				if req.Confirm == nil || !*req.Confirm {
					s.Log.WarnContext(ctx, "menu.DeleteMenu:页面下存在按钮，需要确认删除",
						logger.Int64("menu_id", req.ID),
						logger.Int64("button_count", buttonCount))
					return aurora.NewError(400, "该页面下存在按钮，删除菜单将同时删除所有按钮，请确认")
				}

				// 用户已确认，删除所有按钮及其关联
				var buttons []models.Menu
				if err := tx.Where("parent_id = ? AND type = ?", req.ID, models.MenuTypeButton).
					Find(&buttons).Error; err != nil {
					s.Log.ErrorContext(ctx, "menu.DeleteMenu:查询按钮列表失败", logger.Error(err))
					return aurora.ErrInternal
				}

				for _, btn := range buttons {
					// 删除按钮的角色关联
					if err := tx.Where("menu_id = ?", btn.ID).Delete(&models.RoleMenu{}).Error; err != nil {
						s.Log.ErrorContext(ctx, "menu.DeleteMenu:删除按钮角色关联失败",
							logger.Int64("button_id", btn.ID),
							logger.Error(err))
						return aurora.ErrInternal
					}

					// 删除按钮的API关联
					if err := tx.Where("menu_id = ?", btn.ID).Delete(&models.MenuAPI{}).Error; err != nil {
						s.Log.ErrorContext(ctx, "menu.DeleteMenu:删除按钮API关联失败",
							logger.Int64("button_id", btn.ID),
							logger.Error(err))
						return aurora.ErrInternal
					}

					// 删除按钮
					if err := tx.Delete(&btn).Error; err != nil {
						s.Log.ErrorContext(ctx, "menu.DeleteMenu:删除按钮失败",
							logger.Int64("button_id", btn.ID),
							logger.Error(err))
						return aurora.ErrInternal
					}
				}

				s.Log.InfoContext(ctx, "menu.DeleteMenu:删除页面下的所有按钮",
					logger.Int64("menu_id", req.ID),
					logger.Int64("button_count", buttonCount))
			}
		}

		// 4. 删除菜单的角色关联
		if err := tx.Where("menu_id = ?", req.ID).Delete(&models.RoleMenu{}).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.DeleteMenu:删除菜单角色关联失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 5. 删除菜单的API关联
		if err := tx.Where("menu_id = ?", req.ID).Delete(&models.MenuAPI{}).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.DeleteMenu:删除菜单API关联失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 6. 软删除菜单
		if err := tx.Delete(&menu).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.DeleteMenu:删除菜单失败", logger.Error(err))
			return aurora.ErrInternal
		}

		s.Log.InfoContext(ctx, "menu.DeleteMenu:删除菜单成功",
			logger.Int64("menu_id", req.ID),
			logger.String("title", menu.Title),
			logger.String("type", string(menu.Type)))

		return nil
	})
}

func (s *MenuLogic) UpdateMenu(ctx context.Context, req *types.UpdateMenuReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 检查菜单是否存在
		var menu models.Menu
		if err := tx.Where("id = ?", req.ID).First(&menu).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.Log.WarnContext(ctx, "menu.UpdateMenu:菜单不存在", logger.Int64("menu_id", req.ID))
				return aurora.NewError(404, "菜单不存在")
			}
			s.Log.ErrorContext(ctx, "menu.UpdateMenu:查询菜单失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 2. 验证父级菜单
		if req.ParentID != nil {
			// 不能将自己设置为父级
			if *req.ParentID == req.ID {
				s.Log.WarnContext(ctx, "menu.UpdateMenu:不能将自己设置为父级菜单", logger.Int64("menu_id", req.ID))
				return aurora.NewError(400, "不能将自己设置为父级菜单")
			}

			var parent models.Menu
			if err := tx.Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					s.Log.WarnContext(ctx, "menu.UpdateMenu:父级菜单不存在", logger.Int64("parent_id", *req.ParentID))
					return aurora.NewError(400, "父级菜单不存在")
				}
				s.Log.ErrorContext(ctx, "menu.UpdateMenu:查询父级菜单失败", logger.Error(err))
				return aurora.ErrInternal
			}

			// 按钮不能作为父级菜单
			if parent.IsButton() {
				s.Log.WarnContext(ctx, "menu.UpdateMenu:按钮不能作为父级菜单", logger.Int64("parent_id", *req.ParentID))
				return aurora.NewError(400, "按钮不能作为父级菜单")
			}

			// 检查循环引用：父级菜单不能是当前菜单的子孙节点
			if err := s.checkMenuCircularRef(tx, req.ID, *req.ParentID); err != nil {
				s.Log.WarnContext(ctx, "menu.UpdateMenu:检测到循环引用",
					logger.Int64("menu_id", req.ID),
					logger.Int64("parent_id", *req.ParentID))
				return aurora.NewError(400, "不能设置子孙节点为父级菜单")
			}
		}

		// 3. 验证权限编码唯一性
		if req.PermCode != nil && *req.PermCode != "" && (menu.PermCode == nil || *menu.PermCode != *req.PermCode) {
			var count int64
			if err := tx.Model(&models.Menu{}).
				Where("perm_code = ? AND id != ?", *req.PermCode, req.ID).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "menu.UpdateMenu:查询权限编码失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				s.Log.WarnContext(ctx, "menu.UpdateMenu:权限编码已存在", logger.String("perm_code", *req.PermCode))
				return aurora.NewError(400, "权限编码已存在")
			}
		}

		// 4. 验证路由名称唯一性
		if req.RouteName != nil && *req.RouteName != "" && (menu.RouteName == nil || *menu.RouteName != *req.RouteName) {
			var count int64
			if err := tx.Model(&models.Menu{}).
				Where("route_name = ? AND id != ?", *req.RouteName, req.ID).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "menu.UpdateMenu:查询路由名称失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				s.Log.WarnContext(ctx, "menu.UpdateMenu:路由名称已存在", logger.String("route_name", *req.RouteName))
				return aurora.NewError(400, "路由名称已存在")
			}
		}

		// 5. 更新字段
		updates := make(map[string]interface{})
		if req.ParentID != nil {
			updates["parent_id"] = *req.ParentID
		}
		if req.Type != nil {
			updates["type"] = *req.Type
		}
		if req.PermCode != nil {
			updates["perm_code"] = *req.PermCode
		}
		if req.RouteName != nil {
			updates["route_name"] = *req.RouteName
		}
		if req.Title != nil {
			updates["title"] = *req.Title
		}
		if req.Path != nil {
			updates["path"] = *req.Path
		}
		if req.Component != nil {
			updates["component"] = *req.Component
		}
		if req.Link != nil {
			updates["link"] = *req.Link
		}
		if req.Redirect != nil {
			updates["redirect"] = *req.Redirect
		}
		if req.Icon != nil {
			updates["icon"] = *req.Icon
		}
		if req.ActiveIcon != nil {
			updates["active_icon"] = *req.ActiveIcon
		}
		if req.Sort != nil {
			updates["sort"] = *req.Sort
		}
		if req.Visible != nil {
			updates["visible"] = *req.Visible
		}
		if req.HideChildrenInMenu != nil {
			updates["hide_children_in_menu"] = *req.HideChildrenInMenu
		}
		if req.KeepAlive != nil {
			updates["keep_alive"] = *req.KeepAlive
		}
		if req.AffixTab != nil {
			updates["affix_tab"] = *req.AffixTab
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}

		if len(updates) == 0 {
			s.Log.WarnContext(ctx, "menu.UpdateMenu:没有需要更新的字段", logger.Int64("menu_id", req.ID))
			return aurora.NewError(400, "没有需要更新的字段")
		}

		// 6. 执行更新
		if err := tx.Model(&menu).Updates(updates).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.UpdateMenu:更新菜单失败", logger.Error(err))
			return aurora.ErrInternal
		}

		s.Log.InfoContext(ctx, "menu.UpdateMenu:更新菜单成功",
			logger.Int64("menu_id", req.ID),
			logger.Int("update_fields", len(updates)))

		return nil
	})
}

func (s *MenuLogic) MenuTree(ctx context.Context) ([]*types.MenuTreeResp, error) {
	// 1. 查询所有菜单（不包括按钮）
	var menus []*models.Menu
	if err := s.DB.WithContext(ctx).
		Where("type != ?", models.MenuTypeButton).
		Order("sort ASC, id ASC").
		Find(&menus).Error; err != nil {
		s.Log.ErrorContext(ctx, "menu.MenuTree:查询菜单失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	// 2. 构建并返回树形结构
	return s.buildMenuTree(menus), nil
}

func (s *MenuLogic) MenuDetail(ctx context.Context, req *dto.IDReq) (*types.MenuDetailResp, error) {
	var menu models.Menu
	if err := s.DB.WithContext(ctx).Where("id = ?", req.ID).First(&menu).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.Log.WarnContext(ctx, "menu.MenuDetail:菜单不存在", logger.Int64("menu_id", req.ID))
			return nil, aurora.NewError(404, "菜单不存在")
		}
		s.Log.ErrorContext(ctx, "menu.MenuDetail:查询菜单失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	return &types.MenuDetailResp{
		ID:                 menu.ID,
		ParentID:           menu.ParentID,
		Type:               string(menu.Type),
		PermCode:           menu.PermCode,
		RouteName:          menu.RouteName,
		Title:              menu.Title,
		Path:               menu.Path,
		Component:          menu.Component,
		Link:               menu.Link,
		Redirect:           menu.Redirect,
		Icon:               menu.Icon,
		ActiveIcon:         menu.ActiveIcon,
		Sort:               menu.Sort,
		Visible:            menu.Visible,
		HideChildrenInMenu: menu.HideChildrenInMenu,
		KeepAlive:          menu.KeepAlive,
		AffixTab:           menu.AffixTab,
		Status:             menu.Status,
		CreatedAt:          menu.CreatedAt,
		UpdatedAt:          menu.UpdatedAt,
	}, nil
}

func (s *MenuLogic) CreateButton(ctx context.Context, req *types.CreateButtonReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 验证父级菜单是否存在
		var parent models.Menu
		if err := tx.Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.Log.WarnContext(ctx, "menu.CreateButton:父级菜单不存在", logger.Int64("parent_id", *req.ParentID))
				return aurora.NewError(400, "父级菜单不存在")
			}
			s.Log.ErrorContext(ctx, "menu.CreateButton:查询父级菜单失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 2. 验证父级不能是按钮
		if parent.IsButton() {
			s.Log.WarnContext(ctx, "menu.CreateButton:按钮不能作为父级菜单", logger.Int64("parent_id", *req.ParentID))
			return aurora.NewError(400, "按钮不能作为父级菜单")
		}

		// 3. 验证权限编码唯一性
		var count int64
		if err := tx.Model(&models.Menu{}).
			Where("perm_code = ?", req.PermCode).
			Count(&count).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.CreateButton:查询权限编码失败", logger.Error(err))
			return aurora.ErrInternal
		}
		if count > 0 {
			s.Log.WarnContext(ctx, "menu.CreateButton:权限编码已存在", logger.String("perm_code", req.PermCode))
			return aurora.NewError(400, "权限编码已存在")
		}

		// 4. 创建按钮
		button := &models.Menu{
			ParentID: req.ParentID,
			Type:     models.MenuTypeButton,
			PermCode: &req.PermCode,
			Title:    req.Title,
			Sort:     req.Sort,
			Status:   ptr.Int16Value(req.Status, 1),
		}

		if err := tx.Create(button).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.CreateButton:创建按钮失败", logger.Error(err))
			return aurora.ErrInternal
		}

		s.Log.InfoContext(ctx, "menu.CreateButton:创建按钮成功",
			logger.Int64("button_id", button.ID),
			logger.String("title", button.Title),
			logger.String("perm_code", req.PermCode))

		return nil
	})
}

func (s *MenuLogic) DeleteButton(ctx context.Context, req *types.DeleteButtonReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 检查按钮是否存在
		var button models.Menu
		if err := tx.Where("id = ? AND type = ?", req.ID, models.MenuTypeButton).First(&button).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.Log.WarnContext(ctx, "menu.DeleteButton:按钮不存在", logger.Int64("button_id", req.ID))
				return aurora.NewError(404, "按钮不存在")
			}
			s.Log.ErrorContext(ctx, "menu.DeleteButton:查询按钮失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 2. 删除角色按钮关联
		if err := tx.Where("menu_id = ?", req.ID).Delete(&models.RoleMenu{}).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.DeleteButton:删除角色按钮关联失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 3. 删除按钮API关联
		if err := tx.Where("menu_id = ?", req.ID).Delete(&models.MenuAPI{}).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.DeleteButton:删除按钮API关联失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 4. 软删除按钮
		if err := tx.Delete(&button).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.DeleteButton:删除按钮失败", logger.Error(err))
			return aurora.ErrInternal
		}

		s.Log.InfoContext(ctx, "menu.DeleteButton:删除按钮成功",
			logger.Int64("button_id", req.ID),
			logger.String("title", button.Title))

		return nil
	})
}

func (s *MenuLogic) UpdateButton(ctx context.Context, req *types.UpdateButtonReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 检查按钮是否存在
		var button models.Menu
		if err := tx.Where("id = ? AND type = ?", req.ID, models.MenuTypeButton).First(&button).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				s.Log.WarnContext(ctx, "menu.UpdateButton:按钮不存在", logger.Int64("button_id", req.ID))
				return aurora.NewError(404, "按钮不存在")
			}
			s.Log.ErrorContext(ctx, "menu.UpdateButton:查询按钮失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 2. 验证父级菜单
		if req.ParentID != nil {
			var parent models.Menu
			if err := tx.Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					s.Log.WarnContext(ctx, "menu.UpdateButton:父级菜单不存在", logger.Int64("parent_id", *req.ParentID))
					return aurora.NewError(400, "父级菜单不存在")
				}
				s.Log.ErrorContext(ctx, "menu.UpdateButton:查询父级菜单失败", logger.Error(err))
				return aurora.ErrInternal
			}

			if parent.IsButton() {
				s.Log.WarnContext(ctx, "menu.UpdateButton:按钮不能作为父级菜单", logger.Int64("parent_id", *req.ParentID))
				return aurora.NewError(400, "按钮不能作为父级菜单")
			}
		}

		// 3. 验证权限编码唯一性
		if req.PermCode != nil && *req.PermCode != "" && (button.PermCode == nil || *button.PermCode != *req.PermCode) {
			var count int64
			if err := tx.Model(&models.Menu{}).
				Where("perm_code = ? AND id != ?", *req.PermCode, req.ID).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "menu.UpdateButton:查询权限编码失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				s.Log.WarnContext(ctx, "menu.UpdateButton:权限编码已存在", logger.String("perm_code", *req.PermCode))
				return aurora.NewError(400, "权限编码已存在")
			}
		}

		// 4. 更新字段
		updates := make(map[string]interface{})
		if req.ParentID != nil {
			updates["parent_id"] = *req.ParentID
		}
		if req.PermCode != nil {
			updates["perm_code"] = *req.PermCode
		}
		if req.Title != nil {
			updates["title"] = *req.Title
		}
		if req.Sort != nil {
			updates["sort"] = *req.Sort
		}
		if req.Status != nil {
			updates["status"] = *req.Status
		}

		if len(updates) == 0 {
			s.Log.WarnContext(ctx, "menu.UpdateButton:没有需要更新的字段", logger.Int64("button_id", req.ID))
			return aurora.NewError(400, "没有需要更新的字段")
		}

		// 5. 执行更新
		if err := tx.Model(&button).Updates(updates).Error; err != nil {
			s.Log.ErrorContext(ctx, "menu.UpdateButton:更新按钮失败", logger.Error(err))
			return aurora.ErrInternal
		}

		s.Log.InfoContext(ctx, "menu.UpdateButton:更新按钮成功",
			logger.Int64("button_id", req.ID),
			logger.Int("update_fields", len(updates)))

		return nil
	})
}

func (s *MenuLogic) ButtonList(ctx context.Context) (*types.ButtonListResp, error) {
	var buttons []*models.Menu
	if err := s.DB.WithContext(ctx).
		Where("type = ?", models.MenuTypeButton).
		Order("sort ASC, id ASC").
		Find(&buttons).Error; err != nil {
		s.Log.ErrorContext(ctx, "menu.ButtonList:查询按钮列表失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	items := make([]*types.ButtonItem, len(buttons))
	for i, btn := range buttons {
		permCode := ""
		if btn.PermCode != nil {
			permCode = *btn.PermCode
		}
		items[i] = &types.ButtonItem{
			ID:        btn.ID,
			ParentID:  btn.ParentID,
			PermCode:  permCode,
			Title:     btn.Title,
			Sort:      btn.Sort,
			Status:    btn.Status,
			CreatedAt: btn.CreatedAt,
			UpdatedAt: btn.UpdatedAt,
		}
	}

	return &types.ButtonListResp{Items: items}, nil
}

func (s *MenuLogic) ButtonDetail(ctx context.Context, req *dto.IDReq) (*types.ButtonDetailResp, error) {
	var button models.Menu
	if err := s.DB.WithContext(ctx).
		Where("id = ? AND type = ?", req.ID, models.MenuTypeButton).
		First(&button).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.Log.WarnContext(ctx, "menu.ButtonDetail:按钮不存在", logger.Int64("button_id", req.ID))
			return nil, aurora.NewError(404, "按钮不存在")
		}
		s.Log.ErrorContext(ctx, "menu.ButtonDetail:查询按钮失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	permCode := ""
	if button.PermCode != nil {
		permCode = *button.PermCode
	}

	return &types.ButtonDetailResp{
		ID:        button.ID,
		ParentID:  button.ParentID,
		Type:      string(button.Type),
		PermCode:  permCode,
		Title:     button.Title,
		Sort:      button.Sort,
		Status:    button.Status,
		CreatedAt: button.CreatedAt,
		UpdatedAt: button.UpdatedAt,
	}, nil
}

func (s *MenuLogic) Options(ctx context.Context) ([]*types.MenuOptionsResp, error) {
	// 1. 查询所有菜单（不包括按钮）
	var menus []*models.Menu
	if err := s.DB.WithContext(ctx).
		Select("id, parent_id, title, type").
		Where("type != ?", models.MenuTypeButton).
		Order("sort ASC, id ASC").
		Find(&menus).Error; err != nil {
		s.Log.ErrorContext(ctx, "menu.Options:查询菜单选项失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	// 2. 构建选项树
	return s.buildMenuOptions(menus), nil
}

func NewMenuLogic(ctx *svc.ServiceContext) *MenuLogic {
	return &MenuLogic{ServiceContext: ctx}
}

// validateMenuTypeFields 根据菜单类型验证必填字段
func (s *MenuLogic) validateMenuTypeFields(req *types.CreateMenuReq) error {
	switch models.MenuType(req.Type) {
	case models.MenuTypeDirectory:
		// 目录：必须有路由路径
		if req.Path == nil || *req.Path == "" {
			return errors.New("目录类型菜单必须指定路由路径")
		}

	case models.MenuTypePage:
		// 页面：必须有路由名称、路由路径、组件路径
		if req.RouteName == nil || *req.RouteName == "" {
			return errors.New("页面类型菜单必须指定路由名称")
		}
		if req.Path == nil || *req.Path == "" {
			return errors.New("页面类型菜单必须指定路由路径")
		}
		if req.Component == nil || *req.Component == "" {
			return errors.New("页面类型菜单必须指定组件路径")
		}

	case models.MenuTypeButton:
		// 按钮：必须有权限编码、必须有父级菜单
		if req.PermCode == nil || *req.PermCode == "" {
			return errors.New("按钮类型菜单必须指定权限编码")
		}
		if req.ParentID == nil {
			return errors.New("按钮类型菜单必须指定父级菜单")
		}

	case models.MenuTypeExternal:
		// 外链：必须有外链地址
		if req.Link == nil || *req.Link == "" {
			return errors.New("外链类型菜单必须指定外链地址")
		}
	}

	return nil
}

// checkMenuCircularRef 检查菜单循环引用
func (s *MenuLogic) checkMenuCircularRef(tx *gorm.DB, menuID, newParentID int64) error {
	// 从新父级开始向上查找，如果找到 menuID 则说明有循环引用
	currentID := newParentID
	for currentID != 0 {
		if currentID == menuID {
			return errors.New("检测到循环引用")
		}

		var menu models.Menu
		if err := tx.Select("parent_id").Where("id = ?", currentID).First(&menu).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return err
		}

		if menu.ParentID == nil {
			break
		}
		currentID = *menu.ParentID
	}

	return nil
}

// buildMenuTree 构建菜单树
func (s *MenuLogic) buildMenuTree(menus []*models.Menu) []*types.MenuTreeResp {
	// 构建 map 以便快速查找
	menuMap := make(map[int64]*types.MenuTreeResp)
	var roots []*types.MenuTreeResp

	// 第一遍：创建所有节点
	for _, menu := range menus {
		node := &types.MenuTreeResp{
			ID:                 menu.ID,
			ParentID:           menu.ParentID,
			Type:               string(menu.Type),
			PermCode:           menu.PermCode,
			RouteName:          menu.RouteName,
			Title:              menu.Title,
			Path:               menu.Path,
			Component:          menu.Component,
			Link:               menu.Link,
			Redirect:           menu.Redirect,
			Icon:               menu.Icon,
			ActiveIcon:         menu.ActiveIcon,
			Sort:               menu.Sort,
			Visible:            menu.Visible,
			HideChildrenInMenu: menu.HideChildrenInMenu,
			KeepAlive:          menu.KeepAlive,
			AffixTab:           menu.AffixTab,
			Status:             menu.Status,
			CreatedAt:          menu.CreatedAt,
			Children:           []*types.MenuTreeResp{},
		}
		menuMap[menu.ID] = node
	}

	// 第二遍：建立父子关系
	for _, menu := range menus {
		node := menuMap[menu.ID]
		if menu.ParentID == nil {
			roots = append(roots, node)
		} else {
			parent, exists := menuMap[*menu.ParentID]
			if exists {
				parent.Children = append(parent.Children, node)
			} else {
				// 父节点不存在，作为根节点
				roots = append(roots, node)
			}
		}
	}

	return roots
}

// buildMenuOptions 构建菜单选项树
func (s *MenuLogic) buildMenuOptions(menus []*models.Menu) []*types.MenuOptionsResp {
	// 构建 map 以便快速查找
	optionMap := make(map[int64]*types.MenuOptionsResp)
	var roots []*types.MenuOptionsResp

	// 第一遍：创建所有节点
	for _, menu := range menus {
		node := &types.MenuOptionsResp{
			ID:       menu.ID,
			ParentID: menu.ParentID,
			Title:    menu.Title,
			Type:     string(menu.Type),
			Children: []*types.MenuOptionsResp{},
		}
		optionMap[menu.ID] = node
	}

	// 第二遍：建立父子关系
	for _, menu := range menus {
		node := optionMap[menu.ID]
		if menu.ParentID == nil {
			roots = append(roots, node)
		} else {
			parent, exists := optionMap[*menu.ParentID]
			if exists {
				parent.Children = append(parent.Children, node)
			} else {
				// 父节点不存在，作为根节点
				roots = append(roots, node)
			}
		}
	}

	return roots
}
