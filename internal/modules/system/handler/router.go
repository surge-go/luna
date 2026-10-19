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

	menuGroup := group.Group("/menu")
	menuHandler := NewMenuHandler(logic.NewMenuLogic(ctx))
	{
		// 菜单管理
		menuGroup.POST("/create", aurora.BindE(menuHandler.CreateMenu))
		menuGroup.DELETE("/delete", aurora.BindE(menuHandler.DeleteMenu))
		menuGroup.PUT("/update", aurora.BindE(menuHandler.UpdateMenu))
		menuGroup.GET("/detail", aurora.Bind(menuHandler.MenuDetail))
		menuGroup.GET("/tree", menuHandler.MenuTree)

		// 按钮管理
		menuGroup.POST("/create-button", aurora.BindE(menuHandler.CreateButton))
		menuGroup.DELETE("/delete-button", aurora.BindE(menuHandler.DeleteButton))
		menuGroup.PUT("/update-button", aurora.BindE(menuHandler.UpdateButton))
		menuGroup.GET("/detail-button", aurora.Bind(menuHandler.ButtonDetail))
		menuGroup.GET("/button-list", menuHandler.ButtonList)

		// 选项
		menuGroup.GET("/options", menuHandler.Options)
	}
}
