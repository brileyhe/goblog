package manageController

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/provider"
)

func PluginImportApi(ctx iris.Context) {
	importApi := config.JsonData.PluginImportApi

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": iris.Map{
			"token":      importApi.Token,
			"link_token": importApi.LinkToken,
			"base_url":   config.JsonData.System.BaseUrl,
		},
	})
}

func PluginUpdateApiToken(ctx iris.Context) {
	var req config.PluginImportApiConfig
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	if req.Token != "" {
		config.JsonData.PluginImportApi.Token = req.Token
	}
	if req.LinkToken != "" {
		config.JsonData.PluginImportApi.LinkToken = req.LinkToken
	}
	// 回写
	err := provider.SaveSettingValue(provider.ImportApiSettingKey, config.JsonData.PluginImportApi)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	provider.AddAdminLog(ctx, fmt.Sprintf("更新API导入Token"))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "Token已更新",
	})
}
