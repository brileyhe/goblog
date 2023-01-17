package manageController

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/dao"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/provider"
	"kandaoni.com/anqicms/request"
	"time"
)

func ArchiveList(ctx iris.Context) {
	currentPage := ctx.URLParamIntDefault("current", 1)
	pageSize := ctx.URLParamIntDefault("pageSize", 20)
	categoryId := uint(ctx.URLParamIntDefault("category_id", 0))
	moduleId := uint(ctx.URLParamIntDefault("module_id", 0))
	status := ctx.URLParam("status") // 支持 '':all，draft:0, ok:1, plan:2
	// 回收站
	recycle, _ := ctx.URLParamBool("recycle")
	// 采集的
	collect, _ := ctx.URLParamBool("collect")

	var archives []*model.Archive
	var total int64
	var err error

	var ops func(tx *gorm.DB) *gorm.DB
	if recycle {
		ops = func(tx *gorm.DB) *gorm.DB {
			return tx.Unscoped().Where("`deleted_at` is not null").Order("id desc")
		}
	} else if collect {
		ops = func(tx *gorm.DB) *gorm.DB {
			return tx.Where("`origin_url` != ''").Order("id desc")
		}
	} else {
		// 必须传递分类
		title := ctx.URLParam("title")
		ops = func(tx *gorm.DB) *gorm.DB {
			if moduleId > 0 {
				tx = tx.Where("`module_id` = ?", moduleId)
			}
			if categoryId > 0 {
				tx = tx.Where("`category_id` = ?", categoryId)
			}
			if status == "draft" {
				tx = tx.Where("`status` = ?", config.ContentStatusDraft)
			} else if status == "ok" {
				tx = tx.Where("`status` = ?", config.ContentStatusOK)
			} else if status == "plan" {
				tx = tx.Where("`status` = ?", config.ContentStatusPlan)
			}
			if title != "" {
				tx = tx.Where("`title` like ?", "%"+title+"%")
			}
			tx = tx.Order("id desc")
			return tx
		}
	}
	archives, total, _ = provider.GetArchiveList(ops, currentPage, pageSize)

	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	//读取列表的分类
	categories := provider.GetCacheCategories()
	// 模型
	modules := provider.GetCacheModules()
	for i, v := range archives {
		if v.CategoryId > 0 {
			for _, c := range categories {
				if c.Id == v.CategoryId {
					archives[i].Category = &c
					break
				}
			}
		}
		for _, c := range modules {
			if c.Id == v.ModuleId {
				archives[i].ModuleName = c.Title
			}
		}
	}

	// 给文章生成链接
	for i := range archives {
		archives[i].Link = provider.GetUrl("archive", archives[i], 0)
	}

	ctx.JSON(iris.Map{
		"code":  config.StatusOK,
		"msg":   "",
		"total": total,
		"data":  archives,
	})
}

func ArchiveDetail(ctx iris.Context) {
	id := uint(ctx.URLParamIntDefault("id", 0))

	archive, err := provider.GetArchiveById(id)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	// 读取data
	archive.ArchiveData, err = provider.GetArchiveDataById(archive.Id)
	// 读取 extraDat
	archive.Extra = provider.GetArchiveExtra(archive.ModuleId, archive.Id)

	tags := provider.GetTagsByItemId(archive.Id)
	if len(tags) > 0 {
		var tagNames = make([]string, 0, len(tags))
		for _, v := range tags {
			tagNames = append(tagNames, v.Title)
		}
		archive.Tags = tagNames
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "",
		"data": archive,
	})
}

func ArchiveDetailForm(ctx iris.Context) {
	var req request.Archive
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	// 检查是否有重名
	if !req.ForceSave {
		exists, err := provider.GetArchiveByTitle(req.Title)
		if err == nil && exists.Id != req.Id {
			// 做提示
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  "相同标题的内容已存在",
				"data": exists,
			})
			return
		}
	}
	// 检查 fixed_link
	if req.FixedLink != "" {
		exists, err := provider.GetArchiveByFixedLink(req.FixedLink)
		if err == nil && exists.Id != req.Id {
			// 做提示
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  "已存在相同的固定链接，请更换一个固定链接再提交",
			})
			return
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

	provider.AddAdminLog(ctx, fmt.Sprintf("更新文档：%d => %s", archive.Id, archive.Title))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "文档已更新",
		"data": archive,
	})
}

func ArchiveRecover(ctx iris.Context) {
	var req request.Archive
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	archive, err := provider.GetUnscopedArchiveById(req.Id)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	err = provider.RecoverArchive(archive)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("恢复文档：%d => %s", archive.Id, archive.Title))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "文章已恢复",
	})
}

func ArchiveRelease(ctx iris.Context) {
	var req request.Archive
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	archive, err := provider.GetUnscopedArchiveById(req.Id)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	// 只有待发布的需要发布
	if archive.Status == config.ContentStatusDraft {
		archive.Status = config.ContentStatusOK
		archive.CreatedTime = time.Now().Unix()
		dao.DB.Save(archive)
		err = provider.SuccessReleaseArchive(archive, true)
		if err != nil {
			ctx.JSON(iris.Map{
				"code": config.StatusFailed,
				"msg":  err.Error(),
			})
			return
		}
		provider.AddAdminLog(ctx, fmt.Sprintf("发布文档：%d => %s", archive.Id, archive.Title))
	}

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "文章已发布",
	})
}

func ArchiveDelete(ctx iris.Context) {
	var req request.Archive
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}
	archive, err := provider.GetUnscopedArchiveById(req.Id)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	err = provider.DeleteArchive(archive)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("删除文档：%d => %s", archive.Id, archive.Title))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "文章已删除",
	})
}

func UpdateArchiveRecommend(ctx iris.Context) {
	var req request.ArchivesUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}


	err := provider.UpdateArchiveRecommend(&req)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("批量更新文档Flag：%v => %s", req.Ids, req.Flag))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "文章已更新",
	})
}

func UpdateArchiveStatus(ctx iris.Context) {
	var req request.ArchivesUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}


	err := provider.UpdateArchiveStatus(&req)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("批量更新文档状态：%v => %d", req.Ids, req.Status))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "文章已更新",
	})
}

func UpdateArchiveCategory(ctx iris.Context) {
	var req request.ArchivesUpdateRequest
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}


	err := provider.UpdateArchiveCategory(&req)
	if err != nil {
		ctx.JSON(iris.Map{
			"code": config.StatusFailed,
			"msg":  err.Error(),
		})
		return
	}

	provider.AddAdminLog(ctx, fmt.Sprintf("批量更新文档分类：%v => %d", req.Ids, req.CategoryId))

	ctx.JSON(iris.Map{
		"code": config.StatusOK,
		"msg":  "文章已更新",
	})
}
