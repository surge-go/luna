package handler

import (
	"github.com/surge-go/aurora"
	"github.com/surge-go/luna/internal/modules/system/logic"
	"github.com/surge-go/luna/internal/modules/system/svc"
)

func RegisterRouter(group *aurora.RouterGroup, ctx *svc.ServiceContext) {
	apiGroup := group.Group("/api")

	apiHandler := NewApiHandler(logic.NewApiLogic(ctx))
	{
		// 分组管理
		apiGroup.POST("/create-group", aurora.BindE(apiHandler.CreateGroup))
		apiGroup.DELETE("/delete-group", aurora.BindE(apiHandler.DeleteGroup))
		apiGroup.PUT("/update-group", aurora.BindE(apiHandler.UpdateGroup))
		apiGroup.GET("/detail-group", aurora.Bind(apiHandler.GroupDetail))

		// API 管理
		apiGroup.POST("/create-api", aurora.BindE(apiHandler.CreateApi))
		apiGroup.DELETE("/delete-api", aurora.BindE(apiHandler.DeleteApi))
		apiGroup.PUT("/update-api", aurora.BindE(apiHandler.UpdateApi))
		apiGroup.GET("/detail-api", aurora.Bind(apiHandler.APIDetail))

		// 树形结构和选项
		apiGroup.GET("/tree", apiHandler.Tree)
		apiGroup.GET("/options", apiHandler.Options)
	}
}
