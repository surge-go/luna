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

type ApiLogic struct {
	*svc.ServiceContext
}

func (s *ApiLogic) CreateGroup(ctx context.Context, req *types.ApiCreateGroupReq) error {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		if s.Log != nil {
			s.Log.ErrorContext(ctx, "api.CreateGroup:请求参数验证失败", logger.Error(err))
		}
		return aurora.ErrBadRequest
	}

	// 开启事务
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查Group是否存在
		var count int64
		if err := tx.Model(&models.API{}).Where("group = ?", req.Group).Count(&count).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.CreateGroup:查询分组失败", logger.Error(err))
			return aurora.ErrInternal
		}
		if count > 0 {
			return aurora.ErrConflict.WithMessage("分组标识已存在")
		}

		// 检查同级分组名称是否存在
		if req.ParentID != nil {
			if err := tx.Model(&models.API{}).
				Where("parent_id = ? AND name = ?", *req.ParentID, req.Name).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "api.CreateGroup:查询同级分组名称失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				return aurora.ErrConflict.WithMessage("同级分组名称已存在")
			}
		} else {
			// 检查根级名称
			if err := tx.Model(&models.API{}).
				Where("parent_id IS NULL AND name = ?", req.Name).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "api.CreateGroup:查询根级分组名称失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				return aurora.ErrConflict.WithMessage("根级分组名称已存在")
			}
		}

		// 创建分组
		if err := tx.Create(&models.API{
			ParentID: req.ParentID,
			Group:    ptr.To(req.Group),
			Name:     req.Name,
			Sort:     req.Sort,
			Remark:   ptr.To(req.Remark),
		}).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.CreateGroup:创建分组失败", logger.Error(err))
			return aurora.ErrInternal
		}
		return nil
	})
}

func (s *ApiLogic) DeleteGroup(ctx context.Context, req *dto.IDReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查分组是否存在
		var existGroup models.API
		if err := tx.Where("id = ?", req.ID).First(&existGroup).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return aurora.ErrNotFound.WithMessage("分组不存在")
			}
			s.Log.ErrorContext(ctx, "api.DeleteGroup:查询分组失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 验证是否是分组节点
		if !existGroup.IsGroup() {
			return aurora.ErrBadRequest.WithMessage("该节点不是分组类型")
		}

		// 检查分组下是否有子节点
		var count int64
		if err := tx.Model(&models.API{}).Where("parent_id = ?", req.ID).Count(&count).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.DeleteGroup:查询子节点失败", logger.Error(err))
			return aurora.ErrInternal
		}
		if count > 0 {
			return aurora.ErrConflict.WithMessage("该分组下存在子节点，无法删除")
		}

		// 软删除分组
		if err := tx.Delete(&existGroup).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.DeleteGroup:删除分组失败", logger.Error(err))
			return aurora.ErrInternal
		}

		return nil
	})
}

func (s *ApiLogic) UpdateGroup(ctx context.Context, req *types.ApiUpdateGroupReq) error {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		if s.Log != nil {
			s.Log.ErrorContext(ctx, "api.UpdateGroup:请求参数验证失败", logger.Error(err))
		}
		return aurora.ErrBadRequest
	}

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询当前分组
		var updateGroup models.API
		if err := tx.Where("id = ?", req.ID).First(&updateGroup).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return aurora.ErrNotFound.WithMessage("分组不存在")
			}
			s.Log.ErrorContext(ctx, "api.UpdateGroup:查询分组失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 验证是否是分组节点
		if !updateGroup.IsGroup() {
			return aurora.ErrBadRequest.WithMessage("该节点不是分组类型")
		}

		// 检查分组标识唯一性
		if req.Group != "" && req.Group != updateGroup.GetGroup() {
			var count int64
			if err := tx.Model(&models.API{}).Where("group = ? AND id != ?", req.Group, req.ID).Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "api.UpdateGroup:查询分组标识失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				return aurora.ErrConflict.WithMessage("分组标识已存在")
			}
			updateGroup.Group = ptr.To(req.Group)
		}

		// 检查父级变更
		parentChanged := false
		if req.ParentID != nil {
			// 判断父级是否真的改变了
			if updateGroup.ParentID == nil || *updateGroup.ParentID != *req.ParentID {
				parentChanged = true

				// 检查父级是否存在且是分组类型
				var parent models.API
				if err := tx.Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return aurora.ErrNotFound.WithMessage("父级分组不存在")
					}
					s.Log.ErrorContext(ctx, "api.UpdateGroup:查询父级分组失败", logger.Error(err))
					return aurora.ErrInternal
				}
				if !parent.IsGroup() {
					return aurora.ErrBadRequest.WithMessage("父级节点不是分组类型")
				}

				// 检查是否会造成循环引用（从新父级向上查找）
				if *req.ParentID == req.ID {
					return aurora.ErrBadRequest.WithMessage("不能将分组移动到自己下面")
				}
				checkID := *req.ParentID
				for checkID != 0 {
					if checkID == req.ID {
						return aurora.ErrBadRequest.WithMessage("不能将分组移动到自己的子节点下")
					}
					var node models.API
					if err := tx.Where("id = ?", checkID).First(&node).Error; err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							break
						}
						s.Log.ErrorContext(ctx, "api.UpdateGroup:检查循环引用失败", logger.Error(err))
						return aurora.ErrInternal
					}
					if node.ParentID == nil {
						break
					}
					checkID = *node.ParentID
				}

				updateGroup.ParentID = req.ParentID
			}
		}

		// 检查同级名称唯一性（如果名称或父级发生变化）
		nameChanged := req.Name != "" && req.Name != updateGroup.Name
		if nameChanged || parentChanged {
			checkName := req.Name
			if checkName == "" {
				checkName = updateGroup.Name
			}
			checkParentID := updateGroup.ParentID
			if parentChanged {
				checkParentID = req.ParentID
			}

			var count int64
			query := tx.Model(&models.API{}).Where("name = ? AND id != ?", checkName, req.ID)
			if checkParentID != nil {
				query = query.Where("parent_id = ?", *checkParentID)
			} else {
				query = query.Where("parent_id IS NULL")
			}
			if err := query.Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "api.UpdateGroup:查询同级分组名称失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				return aurora.ErrConflict.WithMessage("同级分组名称已存在")
			}
		}

		// 更新字段
		if req.Name != "" {
			updateGroup.Name = req.Name
		}
		if req.Sort != 0 {
			updateGroup.Sort = req.Sort
		}
		if req.Remark != "" {
			updateGroup.Remark = ptr.To(req.Remark)
		}

		// 执行更新
		if err := tx.Updates(&updateGroup).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.UpdateGroup:更新分组失败", logger.Error(err))
			return aurora.ErrInternal
		}

		return nil
	})
}

func (s *ApiLogic) GroupDetail(ctx context.Context, req *dto.IDReq) (*types.ApiGroupDetailResp, error) {
	var result models.API
	if err := s.DB.WithContext(ctx).Model(&models.API{}).Where("id = ?", req.ID).First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, aurora.ErrNotFound.WithMessage("分组不存在")
		}
		s.Log.ErrorContext(ctx, "api.GroupDetail:查询分组失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	return &types.ApiGroupDetailResp{
		ID:        result.ID,
		ParentID:  result.ParentID,
		Name:      result.Name,
		Sort:      result.Sort,
		Remark:    ptr.From(result.Remark),
		CreatedAt: result.CreatedAt,
	}, nil
}

func (s *ApiLogic) CreateApi(ctx context.Context, req *types.ApiCreateAPIReq) error {
	if err := req.Validate(); err != nil {
		s.Log.ErrorContext(ctx, "api.CreateApi:校验请求参数失败", logger.Error(err))
		return aurora.ErrBadRequest.WithMessage(err.Error())
	}

	// API 必须有父级分组
	if req.ParentID == nil {
		return aurora.ErrBadRequest.WithMessage("API必须指定父级分组")
	}

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查父分组是否存在且是分组类型
		var parent models.API
		if err := tx.Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return aurora.ErrNotFound.WithMessage("父级分组不存在")
			}
			s.Log.ErrorContext(ctx, "api.CreateApi:查询父级分组失败", logger.Error(err))
			return aurora.ErrInternal
		}
		if !parent.IsGroup() {
			return aurora.ErrBadRequest.WithMessage("父级节点不是分组类型")
		}

		// 检查 Method + Path 唯一性（全局）
		var count int64
		if err := tx.Model(&models.API{}).
			Where("method = ? AND path = ?", req.Method, req.Path).
			Count(&count).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.CreateApi:查询API路径失败", logger.Error(err))
			return aurora.ErrInternal
		}
		if count > 0 {
			return aurora.ErrConflict.WithMessage("该API路径已存在")
		}

		// 检查同级名称唯一性
		if err := tx.Model(&models.API{}).
			Where("name = ? AND parent_id = ?", req.Name, *req.ParentID).
			Count(&count).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.CreateApi:查询同级API名称失败", logger.Error(err))
			return aurora.ErrInternal
		}
		if count > 0 {
			return aurora.ErrConflict.WithMessage("同级API名称已存在")
		}

		// 创建API节点
		api := &models.API{
			ParentID: req.ParentID,
			Name:     req.Name,
			Method:   ptr.To(req.Method),
			Path:     ptr.To(req.Path),
			Sort:     req.Sort,
		}
		if req.Remark != "" {
			api.Remark = ptr.To(req.Remark)
		}

		// 执行创建
		if err := tx.Create(api).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.CreateApi:创建API失败", logger.Error(err))
			return aurora.ErrInternal
		}

		return nil
	})
}

func (s *ApiLogic) DeleteApi(ctx context.Context, req *types.ApiDeleteAPIReq) error {
	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查API是否存在
		var api models.API
		if err := tx.Where("id = ?", req.ID).First(&api).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return aurora.ErrNotFound.WithMessage("API不存在")
			}
			s.Log.ErrorContext(ctx, "api.DeleteApi:查询API失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 验证是否是API节点（而非分组节点）
		if api.IsGroup() {
			return aurora.ErrBadRequest.WithMessage("该节点是分组类型，请使用分组删除接口")
		}

		// 检查是否与菜单绑定
		var count int64
		if err := tx.Model(&models.MenuAPI{}).Where("api_id = ?", req.ID).Count(&count).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.DeleteApi:查询菜单绑定失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 如果有菜单绑定且未确认，拒绝删除
		if count > 0 && !req.Confirm {
			return aurora.ErrConflict.WithMessage("该API已被菜单绑定，请确认后删除")
		}

		// 先删除菜单绑定关系
		if count > 0 {
			if err := tx.Where("api_id = ?", req.ID).Delete(&models.MenuAPI{}).Error; err != nil {
				s.Log.ErrorContext(ctx, "api.DeleteApi:删除菜单绑定失败", logger.Error(err))
				return aurora.ErrInternal
			}
		}

		// 软删除API
		if err := tx.Delete(&api).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.DeleteApi:删除API失败", logger.Error(err))
			return aurora.ErrInternal
		}

		return nil
	})
}

func (s *ApiLogic) UpdateApi(ctx context.Context, req *types.ApiUpdateAPIReq) error {
	// 验证请求参数
	if err := req.Validate(); err != nil {
		s.Log.ErrorContext(ctx, "api.UpdateApi:校验请求参数失败", logger.Error(err))
		return aurora.ErrBadRequest.WithMessage(err.Error())
	}

	return s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询当前API
		var api models.API
		if err := tx.Where("id = ?", req.ID).First(&api).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return aurora.ErrNotFound.WithMessage("API不存在")
			}
			s.Log.ErrorContext(ctx, "api.UpdateApi:查询API失败", logger.Error(err))
			return aurora.ErrInternal
		}

		// 验证是否是API节点
		if api.IsGroup() {
			return aurora.ErrBadRequest.WithMessage("该节点是分组类型，请使用分组更新接口")
		}

		// 检查 Method+Path 唯一性（如果有变更）
		checkMethod := api.GetMethod()
		checkPath := api.GetPath()

		methodChanged := req.Method != "" && req.Method != checkMethod
		pathChanged := req.Path != "" && req.Path != checkPath

		if methodChanged || pathChanged {
			if req.Method != "" {
				checkMethod = req.Method
			}
			if req.Path != "" {
				checkPath = req.Path
			}

			var count int64
			if err := tx.Model(&models.API{}).
				Where("method = ? AND path = ? AND id != ?", checkMethod, checkPath, req.ID).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "api.UpdateApi:查询API路径失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				return aurora.ErrConflict.WithMessage("该API路径已存在")
			}
		}

		// 检查父级变更
		parentChanged := false
		if req.ParentID != nil {
			if api.ParentID == nil || *api.ParentID != *req.ParentID {
				parentChanged = true

				// 检查父级是否存在且是分组类型
				var parent models.API
				if err := tx.Where("id = ?", *req.ParentID).First(&parent).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return aurora.ErrNotFound.WithMessage("父级分组不存在")
					}
					s.Log.ErrorContext(ctx, "api.UpdateApi:查询父级分组失败", logger.Error(err))
					return aurora.ErrInternal
				}
				if !parent.IsGroup() {
					return aurora.ErrBadRequest.WithMessage("父级节点不是分组类型")
				}

				api.ParentID = req.ParentID
			}
		}

		// 检查同级名称唯一性（如果名称或父级发生变化）
		nameChanged := req.Name != "" && req.Name != api.Name
		if nameChanged || parentChanged {
			checkName := req.Name
			if checkName == "" {
				checkName = api.Name
			}
			checkParentID := api.ParentID
			if parentChanged {
				checkParentID = req.ParentID
			}

			var count int64
			if err := tx.Model(&models.API{}).
				Where("name = ? AND parent_id = ? AND id != ?", checkName, checkParentID, req.ID).
				Count(&count).Error; err != nil {
				s.Log.ErrorContext(ctx, "api.UpdateApi:查询同级API名称失败", logger.Error(err))
				return aurora.ErrInternal
			}
			if count > 0 {
				return aurora.ErrConflict.WithMessage("同级API名称已存在")
			}
		}

		// 更新字段
		if req.Name != "" {
			api.Name = req.Name
		}
		if req.Method != "" {
			api.Method = ptr.To(req.Method)
		}
		if req.Path != "" {
			api.Path = ptr.To(req.Path)
		}
		if req.Sort != 0 {
			api.Sort = req.Sort
		}
		if req.Remark != "" {
			api.Remark = ptr.To(req.Remark)
		}

		// 执行更新
		if err := tx.Updates(&api).Error; err != nil {
			s.Log.ErrorContext(ctx, "api.UpdateApi:更新API失败", logger.Error(err))
			return aurora.ErrInternal
		}

		return nil
	})
}

func (s *ApiLogic) APIDetail(ctx context.Context, req *dto.IDReq) (*types.ApiAPIDetailResp, error) {
	// 验证请求参数
	if req.ID <= 0 {
		return nil, aurora.ErrBadRequest.WithMessage("无效的API ID")
	}

	// 查询API详情
	var api models.API
	if err := s.DB.WithContext(ctx).Where("id = ?", req.ID).First(&api).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, aurora.ErrNotFound.WithMessage("API不存在")
		}
		s.Log.ErrorContext(ctx, "api.APIDetail:查询API详情失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	// 验证是否是API节点
	if api.IsGroup() {
		return nil, aurora.ErrBadRequest.WithMessage("该节点是分组类型，请使用分组详情接口")
	}

	// 构建响应
	resp := &types.ApiAPIDetailResp{
		ID:        api.ID,
		ParentID:  api.ParentID,
		Name:      api.Name,
		Method:    api.GetMethod(),
		Path:      api.GetPath(),
		Sort:      api.Sort,
		Remark:    ptr.From(api.Remark),
		CreatedAt: api.CreatedAt,
	}

	return resp, nil
}

func (s *ApiLogic) Tree(ctx context.Context) ([]*types.ApiTreeResp, error) {
	// 查询所有节点（包括分组和API）
	var nodes []*models.API
	if err := s.DB.WithContext(ctx).
		Order("sort ASC, id ASC").
		Find(&nodes).Error; err != nil {
		s.Log.ErrorContext(ctx, "api.Tree:查询API树失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	// 使用 map 构建树（两次遍历，O(n) 复杂度）
	return s.buildTree(nodes), nil
}

// buildTree 构建树形结构（两次遍历算法）
func (s *ApiLogic) buildTree(nodes []*models.API) []*types.ApiTreeResp {
	// 第一次遍历：构建所有分组节点的映射
	groupMap := make(map[int64]*types.ApiTreeResp)
	var roots []*types.ApiTreeResp

	// 先创建所有分组节点
	for _, node := range nodes {
		if node.IsGroup() {
			treeNode := &types.ApiTreeResp{
				ID:        node.ID,
				ParentID:  node.ParentID,
				Group:     node.Group,
				Name:      node.Name,
				Sort:      node.Sort,
				Remark:    node.Remark,
				CreatedAt: node.CreatedAt,
				Children:  make([]*types.ApiTreeResp, 0),
				Apis:      make([]*types.ApiTreeNode, 0),
			}
			groupMap[node.ID] = treeNode

			// 根节点收集
			if node.ParentID == nil {
				roots = append(roots, treeNode)
			}
		}
	}

	// 第二次遍历：建立父子关系
	for _, node := range nodes {
		if node.ParentID == nil {
			continue // 根节点已处理
		}

		parent, exists := groupMap[*node.ParentID]
		if !exists {
			continue // 父节点不存在（数据不一致）
		}

		if node.IsGroup() {
			// 子分组追加到 Children
			childGroup := groupMap[node.ID]
			parent.Children = append(parent.Children, childGroup)
		} else {
			// API 节点追加到 Apis
			parent.Apis = append(parent.Apis, &types.ApiTreeNode{
				ID:        node.ID,
				ParentID:  *node.ParentID,
				Name:      node.Name,
				Sort:      node.Sort,
				Method:    node.Method,
				Path:      node.Path,
				Remark:    node.Remark,
				CreatedAt: node.CreatedAt,
			})
		}
	}

	return roots
}

func (s *ApiLogic) Options(ctx context.Context) ([]*types.ApiOptionsResp, error) {
	// 查询所有节点（包括分组和API）
	var nodes []*models.API
	if err := s.DB.WithContext(ctx).
		Select("id, parent_id, group, name").
		Order("sort ASC, id ASC").
		Find(&nodes).Error; err != nil {
		s.Log.ErrorContext(ctx, "api.Options:查询API选项失败", logger.Error(err))
		return nil, aurora.ErrInternal
	}

	// 使用 map 构建选项树
	return s.buildOptions(nodes), nil
}

// buildOptions 构建选项树结构
func (s *ApiLogic) buildOptions(nodes []*models.API) []*types.ApiOptionsResp {
	// 第一次遍历：构建所有分组节点的映射
	groupMap := make(map[int64]*types.ApiOptionsResp)
	var roots []*types.ApiOptionsResp

	// 先创建所有分组节点
	for _, node := range nodes {
		if node.IsGroup() {
			optNode := &types.ApiOptionsResp{
				ID:       node.ID,
				ParentID: node.ParentID,
				Group:    node.Group,
				Name:     node.Name,
				Children: make([]*types.ApiOptionsResp, 0),
				Apis:     make([]*types.ApiOptionsNode, 0),
			}
			groupMap[node.ID] = optNode

			// 根节点收集
			if node.ParentID == nil {
				roots = append(roots, optNode)
			}
		}
	}

	// 第二次遍历：建立父子关系
	for _, node := range nodes {
		if node.ParentID == nil {
			continue // 根节点已处理
		}

		parent, exists := groupMap[*node.ParentID]
		if !exists {
			continue // 父节点不存在
		}

		if node.IsGroup() {
			// 子分组追加到 Children
			childGroup := groupMap[node.ID]
			parent.Children = append(parent.Children, childGroup)
		} else {
			// API 节点追加到 Apis
			parent.Apis = append(parent.Apis, &types.ApiOptionsNode{
				ID:       node.ID,
				ParentID: *node.ParentID,
				Name:     node.Name,
			})
		}
	}

	return roots
}

func NewApiLogic(ctx *svc.ServiceContext) *ApiLogic {
	return &ApiLogic{ctx}
}
