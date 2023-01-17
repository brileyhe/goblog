package manageController

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/provider"
	"kandaoni.com/anqicms/request"
	"strings"
)

func PluginGuestbookList(ctx iris.Context) {
	//需要支持分页，还要支持搜索
	currentPage := ctx.URLParamIntDefault("current", 1)
	pageSize := ctx.URLParamIntDefault("pageSize", 20)
	keyword := ctx.URLParam("keyword")

	guestbookList, total, err := provider.GetGuestbookList(keyword, currentPage, pageSize)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  "",
		})
		return
	}

	ctx.JSON(iris.Map{
		"code":  config.StatusOK,
		"msg":   "",
		"total": total,
		"data":  guestbookList,
	})
}

func PluginGuestbookDelete(ctx iris.Context) {
	var req request.PluginGuestbookDelete
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	if req.Id > 0 {
		//删一条
		guestbook, err := provider.GetGuestbookById(req.Id)
		if err != nil {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  err.Error(),
			})
			return
		}

		err = provider.DeleteGuestbook(guestbook)
		if err != nil {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  err.Error(),
			})
			return
		}
	} else if len(req.Ids) > 0 {
		//删除多条
		for _, id := range req.Ids {
			guestbook, err := provider.GetGuestbookById(id)
			if err != nil {
				continue
			}

			_ = provider.DeleteGuestbook(guestbook)
		}
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("删除留言：%d, %v", req.Id, req.Ids))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "已执行删除操作",
	})
}

func PluginGuestbookExport(ctx iris.Context) {
	guestbooks, err := provider.GetAllGuestbooks()
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	fields := config.GetGuestbookFields()
	//header
	var header []string
	for _, v := range fields {
		header = append(header, v.Name)
	}

	var content [][]interface{}
	//content
	for _, v := range guestbooks {
		var item []interface{}
		for _, f := range fields {
			if f.IsSystem {
				if f.FieldName == "user_name" {
					item = append(item, v.UserName)
				} else if f.FieldName == "contact" {
					item = append(item, v.Contact)
				} else if f.FieldName == "content" {
					item = append(item, v.Content)
				} else {
					item = append(item, "")
				}
			} else {
				item = append(item, v.ExtraData[f.Name])
			}
		}

		content = append(content, item)
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("导出留言"))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": iris.Map{
			"header":  header,
			"content": content,
		},
	})
}

func PluginGuestbookSetting(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": iris.Map{
			"return_message": config.JsonData.PluginGuestbook.ReturnMessage,
			"fields":         config.GetGuestbookFields(),
		},
	})
}

func PluginGuestbookSettingForm(ctx iris.Context) {
	var req config.PluginGuestbookConfig
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	var fields []*config.CustomField
	var existsFields = map[string]struct{}{}
	for _, v := range req.Fields {
		if !v.IsSystem {
			if v.FieldName == "" {
				v.FieldName = strings.ReplaceAll(library.GetPinyin(v.Name), "-", "_")
			}
		}
		if _, ok := existsFields[v.FieldName]; !ok {
			existsFields[v.FieldName] = struct{}{}
			fields = append(fields, v)
		}
	}

	config.JsonData.PluginGuestbook.ReturnMessage = req.ReturnMessage
	config.JsonData.PluginGuestbook.Fields = fields

	err := provider.SaveSettingValue(provider.GuestbookSettingKey, config.JsonData.PluginGuestbook)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("修改留言设置信息"))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "配置已更新",
	})
}
