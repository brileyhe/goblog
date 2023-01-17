package controller

import (
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/dao"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/provider"
	"kandaoni.com/anqicms/request"
	"kandaoni.com/anqicms/response"
	"strconv"
	"strings"
	"time"
)

func ApiImportArchive(ctx iris.Context) {
	id := uint(ctx.PostValueIntDefault("id", 0))
	title := ctx.PostValueTrim("title")
	seoTitle := ctx.PostValueTrim("seo_title")
	content := ctx.PostValueTrim("content")
	categoryId := uint(ctx.PostValueIntDefault("category_id", 0))
	keywords := ctx.PostValueTrim("keywords")
	description := ctx.PostValueTrim("description")
	logo := ctx.PostValueTrim("logo")
	publishTime := ctx.PostValueTrim("publish_time")
	tmpTag := ctx.PostValueTrim("tag")
	images, _ := ctx.PostValues("images[]")
	urlToken := ctx.PostValueTrim("url_token")
	draft, _ := ctx.PostValueBool("draft")
	cover, _ := ctx.PostValueBool("cover")
	template := ctx.PostValueTrim("template")
	canonicalUrl := ctx.PostValueTrim("canonical_url")
	fixedLink := ctx.PostValueTrim("fixed_link")
	flag := ctx.PostValueTrim("flag")
	price := ctx.PostValueInt64Default("price", 0)
	stock := ctx.PostValueInt64Default("stock", 0)
	readLevel := ctx.PostValueIntDefault("read_level", 0)

	category := provider.GetCategoryFromCache(categoryId)
	if category == nil || category.Type != config.CategoryTypeArchive {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("请选择一个栏目"),
		})
		return
	}
	module := provider.GetModuleFromCache(category.ModuleId)
	if module == nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("未定义模型"),
		})
		return
	}

	if title == "" {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("请填写文章标题"),
		})
		return
	}
	if content == "" {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("请填写文章内容"),
		})
		return
	}

	var req = request.Archive{
		Title:        title,
		SeoTitle:     seoTitle,
		CategoryId:   categoryId,
		Keywords:     keywords,
		Description:  description,
		Content:      content,
		Template:     template,
		CanonicalUrl: canonicalUrl,
		FixedLink:    fixedLink,
		Flag:         flag,
		Price:        price,
		Stock:        stock,
		ReadLevel:    readLevel,
		Images:       images,
		UrlToken:     urlToken,
		Extra:        map[string]interface{}{},
		Draft:        draft,
	}

	// 如果传了ID，则采用覆盖的形式
	if id > 0 {
		_, err := provider.GetArchiveById(id)
		if err != nil {
			// 不存在，创建一个
			archive := model.Archive{
				Title:       title,
				SeoTitle:    seoTitle,
				UrlToken:    urlToken,
				Keywords:    keywords,
				Description: description,
				ModuleId:    category.ModuleId,
				CategoryId:  categoryId,
				Status:      0,
				Logo:        logo,
			}
			archive.Id = id
			err = dao.DB.Create(&archive).Error

			if err != nil {
				ctx.JSON(iris.Map{
					"code": config.StatusFailed,
					"msg":  config.Lang("导入文章失败"),
				})
				return
			}
			req.Id = id
		} else {
			// 已存在
			if cover {
				req.Id = id
			} else {
				ctx.JSON(iris.Map{
					"code": config.StatusFailed,
					"msg":  config.Lang("文档ID重复，不允许重复导入"),
				})
				return
			}
		}
	} else {
		// 标题重复的不允许导入
		exists, err := provider.GetArchiveByTitle(title)
		if err == nil {
			if cover {
				req.Id = exists.Id
			} else {
				ctx.JSON(iris.Map{
					"code": config.StatusFailed,
					"msg":  config.Lang("文档标题重复，不允许重复导入"),
				})
				return
			}
		}
	}

	if publishTime != "" {
		timeStamp, err := time.Parse("2006-01-02 15:04:05", publishTime)
		if err == nil {
			req.CreatedTime = timeStamp.Unix()
		}
	}
	if logo != "" {
		req.Images = append(req.Images, logo)
	}
	if tmpTag != "" {
		tags := strings.Split(strings.ReplaceAll(tmpTag, "，", ","), ",")
		req.Tags = tags
	}

	// 处理extraFields
	if len(module.Fields) > 0 {
		for _, v := range module.Fields {
			if v.Type == config.CustomFieldTypeCheckbox {
				// 多选值
				value, _ := ctx.PostValues(v.FieldName)
				if len(value) > 0 {
					req.Extra[v.FieldName] = map[string]interface{}{
						"value": value,
					}
				}
			} else if v.Type == config.CustomFieldTypeNumber {
				value := ctx.PostValueIntDefault(v.FieldName, 0)
				if value > 0 {
					req.Extra[v.FieldName] = map[string]interface{}{
						"value": value,
					}
				}
			} else {
				value := ctx.PostValue(v.FieldName)
				if value != "" {
					req.Extra[v.FieldName] = map[string]interface{}{
						"value": value,
					}
				}
			}
		}
	}

	archive, err := provider.SaveArchive(&req)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  config.Lang("发布成功"),
		"data": iris.Map{
			"url": provider.GetUrl("archive", archive, 0),
			"id":  archive.Id,
		},
	})
}

func ApiImportGetCategories(ctx iris.Context) {
	tmpModuleId := ctx.FormValue("module_id")
	moduleId, _ := strconv.Atoi(tmpModuleId)

	module := provider.GetModuleFromCache(uint(moduleId))

	if module == nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("未定义模型"),
		})
		return
	}

	tmpCategories, _ := provider.GetCategories(uint(moduleId), "", 0)

	var categories []response.ApiCategory
	for i := range tmpCategories {
		categories = append(categories, response.ApiCategory{
			Id:       tmpCategories[i].Id,
			ParentId: tmpCategories[i].ParentId,
			Title:    tmpCategories[i].Title,
		})
	}

	ctx.JSON(iris.Map{
		"code": config.StatusApiSuccess,
		"msg":  config.Lang("获取成功"),
		"data": categories,
	})
}

func ApiImportCreateFriendLink(ctx iris.Context) {
	// 增加支持 didi 友链的批量导入
	form := library.NewForm(ctx.Request().Form)
	var otherList []map[string]string
	err := form.Bind(&otherList, "other_list")
	if err == nil && len(otherList) > 0 {
		for _, item := range otherList {
			friendLink, err := provider.GetLinkByLink(item["url"])
			if err != nil {
				friendLink = &model.Link{}
			}
			friendLink.Title = item["name"]
			friendLink.Link = item["url"]
			friendLink.Contact = item["qq"]
			friendLink.Status = 0
			friendLink.Save(dao.DB)
		}

		provider.DeleteCacheIndex()

		ctx.JSON(iris.Map{
			"code": config.StatusOK,
			"msg":  config.Lang("链接已保存"),
		})
		return
	}

	title := ctx.PostValueTrim("title")
	link := ctx.PostValueTrim("link")
	if linkUrl := ctx.PostValueTrim("url"); linkUrl != "" {
		link = linkUrl
	}
	nofollow := uint(ctx.PostValueIntDefault("nofollow", 0))
	backLink := ctx.PostValueTrim("back_link")
	myTitle := ctx.PostValueTrim("my_title")
	myLink := ctx.PostValueTrim("my_link")
	contact := ctx.PostValueTrim("contact")
	if qq := ctx.PostValueTrim("qq"); qq != "" {
		contact = qq
	}
	if email := ctx.PostValueTrim("email"); email != "" {
		contact = email
	}
	remark := ctx.PostValueTrim("remark")

	friendLink, err := provider.GetLinkByLink(link)
	if err != nil {
		friendLink = &model.Link{
			Status: 0,
		}
	}

	friendLink.Title = title
	friendLink.Link = link
	if backLink != "" {
		friendLink.BackLink = backLink
	}
	if myTitle != "" {
		friendLink.MyTitle = myTitle
	}
	if myLink != "" {
		friendLink.MyLink = myLink
	}
	if contact != "" {
		friendLink.Contact = contact
	}
	if remark != "" {
		friendLink.Remark = remark
	}
	friendLink.Nofollow = nofollow
	friendLink.Status = 0

	err = friendLink.Save(dao.DB)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	// 保存完毕，实时监测
	go provider.PluginLinkCheck(friendLink)

	provider.DeleteCacheIndex()

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  config.Lang("链接已保存"),
	})
}

func ApiImportDeleteFriendLink(ctx iris.Context) {
	link := ctx.PostValueTrim("link")
	if linkUrl := ctx.PostValueTrim("url"); linkUrl != "" {
		link = linkUrl
	}

	if link == "" {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("link必填"),
		})
		return
	}

	friendLink, err := provider.GetLinkByLink(link)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("链接不存在"),
		})
		return
	}

	err = friendLink.Delete(dao.DB)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	provider.DeleteCacheIndex()

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  config.Lang("链接已删除"),
	})
}

func ApiImportGetFriendLinks(ctx iris.Context) {
	links, _ := provider.GetLinkList()
	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": links,
	})
}

func ApiImportCheckFriendLink(ctx iris.Context) {
	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  config.Lang("验证成功"),
	})
}

func VerifyApiToken(ctx iris.Context) {
	token := ctx.FormValue("token")
	if token != config.JsonData.PluginImportApi.Token {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("Token错误"),
		})
		return
	}

	ctx.Next()
}

func VerifyApiLinkToken(ctx iris.Context) {
	token := ctx.FormValue("token")
	if didiToken := ctx.GetHeader("didi-token"); didiToken != "" {
		token = didiToken
	}
	if token != config.JsonData.PluginImportApi.LinkToken {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("Token错误"),
		})
		return
	}

	ctx.Next()
}

func CheckApiOpen(ctx iris.Context) {
	if 1 != config.JsonData.Safe.APIOpen {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  config.Lang("API接口功能未开放"),
		})
		return
	}

	ctx.Next()
}
