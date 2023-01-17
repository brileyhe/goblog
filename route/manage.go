package route

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/controller"
	"kandaoni.com/anqicms/controller/manageController"
	"kandaoni.com/anqicms/middleware"
)

func manageRoute(app *iris.Application) {
	system := app.Party("/system", manageController.AdminFileServ)
	{
		system.HandleDir("/", fmt.Sprintf("%ssystem", config.ExecPath))
	}
	manage := system.Party("/api", middleware.ParseAdminUrl)
	{
		manage.Post("/login", manageController.AdminLogin)
		manage.Get("/captcha", controller.GenerateCaptcha)

		version := manage.Party("/version")
		{
			version.Get("/info", manageController.Version)
			version.Get("/check", manageController.CheckVersion)
			version.Post("/upgrade", manageController.VersionUpgrade)
		}

		anqi := manage.Party("/anqi", middleware.ParseAdminToken)
		{
			anqi.Get("/info", manageController.GetAnqiInfo)
			anqi.Post("/login", manageController.AnqiLogin)
			anqi.Post("/upload", manageController.AnqiUploadAttachment)
			anqi.Post("/template/share", manageController.AnqiShareTemplate)
			anqi.Post("/template/download", manageController.AnqiDownloadTemplate)
			anqi.Post("/feedback", manageController.AnqiSendFeedback)
			anqi.Post("/restart", manageController.RestartAnqicms)
		}

		admin := manage.Party("/admin", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			admin.Get("/menus", manageController.AdminMenus)
			admin.Get("/list", manageController.AdminList)
			admin.Get("/detail", manageController.AdminDetail)
			admin.Post("/detail", manageController.AdminDetailForm)
			admin.Post("/delete", manageController.AdminDetailDelete)
			admin.Post("/logout", manageController.AdminLogout)
			admin.Get("/logs/login", manageController.GetAdminLoginLog)
			admin.Get("/logs/action", manageController.GetAdminLog)
			admin.Get("/group/list", manageController.AdminGroupList)
			admin.Get("/group/detail", manageController.AdminGroupDetail)
			admin.Post("/group/detail", manageController.AdminGroupDetailForm)
			admin.Post("/group/delete", manageController.AdminGroupDelete)
		}

		setting := manage.Party("/setting", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			setting.Get("/system", manageController.SettingSystem)
			setting.Get("/content", manageController.SettingContent)
			setting.Get("/index", manageController.SettingIndex)
			setting.Get("/nav", manageController.SettingNav)
			setting.Get("/nav/type", manageController.SettingNavType)
			setting.Get("/contact", manageController.SettingContact)
			setting.Get("/cache", manageController.SettingCache)
			setting.Get("/safe", manageController.SettingSafe)

			setting.Post("/system", manageController.SettingSystemForm)
			setting.Post("/content", manageController.SettingContentForm)
			setting.Post("/thumb/rebuild", manageController.SettingThumbRebuild)
			setting.Post("/index", manageController.SettingIndexForm)
			setting.Post("/nav", manageController.SettingNavForm)
			setting.Post("/nav/delete", manageController.SettingNavDelete)
			setting.Post("/nav/type", manageController.SettingNavTypeForm)
			setting.Post("/nav/type/delete", manageController.SettingNavTypeDelete)
			setting.Post("/contact", manageController.SettingContactForm)
			setting.Post("/cache", manageController.SettingCacheForm)
			setting.Post("/convert/webp", manageController.ConvertImageToWebp)
			setting.Post("/safe", manageController.SettingSafeForm)

		}

		collector := manage.Party("/collector", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			//采集全局设置
			collector.Get("/setting", manageController.HandleCollectSetting)
			collector.Post("/setting", manageController.HandleSaveCollectSetting)
			//批量替换文章内容
			collector.Post("/article/replace", manageController.HandleReplaceArticles)
			collector.Post("/article/pseudo", manageController.HandleArticlePseudo)
			collector.Post("/article/collect", manageController.HandleArticleCollect)
			collector.Post("/keyword/dig", manageController.HandleDigKeywords)
		}

		attachment := manage.Party("/attachment", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			attachment.Get("/list", manageController.AttachmentList)
			attachment.Post("/upload", manageController.AttachmentUpload)
			attachment.Post("/delete", manageController.AttachmentDelete)
			attachment.Post("/edit", manageController.AttachmentEdit)

			attachment.Post("/category", manageController.AttachmentChangeCategory)
			attachment.Get("/category/list", manageController.AttachmentCategoryList)
			attachment.Post("/category/detail", manageController.AttachmentCategoryDetailForm)
			attachment.Post("/category/delete", manageController.AttachmentCategoryDelete)
		}

		module := manage.Party("/module", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			module.Get("/list", manageController.ModuleList)
			module.Get("/detail", manageController.ModuleDetail)
			module.Post("/detail", manageController.ModuleDetailForm)
			module.Post("/delete", manageController.ModuleDelete)
			module.Post("/field/delete", manageController.ModuleFieldsDelete)
		}

		category := manage.Party("/category", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			category.Get("/list", manageController.CategoryList)
			category.Get("/detail", manageController.CategoryDetail)
			category.Post("/detail", manageController.CategoryDetailForm)
			category.Post("/delete", manageController.CategoryDelete)
		}

		archive := manage.Party("/archive", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			archive.Get("/list", manageController.ArchiveList)
			archive.Get("/detail", manageController.ArchiveDetail)
			archive.Post("/detail", manageController.ArchiveDetailForm)
			archive.Post("/delete", manageController.ArchiveDelete)
			archive.Post("/recover", manageController.ArchiveRecover)
			archive.Post("/release", manageController.ArchiveRelease)
			archive.Post("/recommend", manageController.UpdateArchiveRecommend)
			archive.Post("/status", manageController.UpdateArchiveStatus)
			archive.Post("/category", manageController.UpdateArchiveCategory)
		}

		statistic := manage.Party("/statistic", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			statistic.Get("/spider", manageController.StatisticSpider)
			statistic.Get("/traffic", manageController.StatisticTraffic)
			statistic.Get("/detail", manageController.StatisticDetail)
			statistic.Get("/include", manageController.GetSpiderInclude)
			statistic.Get("/include/detail", manageController.GetSpiderIncludeDetail)
			statistic.Get("/summary", manageController.GetStatisticsSummary)
			statistic.Get("/dashboard", manageController.GetStatisticsDashboard)
		}

		design := manage.Party("/design", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			design.Get("/list", manageController.GetDesignList)
			design.Get("/info", manageController.GetDesignInfo)
			design.Post("/save", manageController.SaveDesignInfo)
			design.Post("/delete", manageController.DeleteDesignInfo)
			design.Post("/download", manageController.DownloadDesignInfo)
			design.Post("/upload", manageController.UploadDesignInfo)
			design.Post("/use", manageController.UseDesignInfo)
			design.Post("/data/restore", manageController.RestoreDesignData)
			design.Post("/data/backup", manageController.BackupDesignData)
			design.Get("/file/info", manageController.GetDesignFileDetail)
			design.Get("/file/histories", manageController.GetDesignFileHistories)
			design.Post("/file/history/delete", manageController.DeleteDesignFileHistories)
			design.Post("/file/restore", manageController.RestoreDesignFile)
			design.Post("/file/save", manageController.SaveDesignFile)
			design.Post("/file/copy", manageController.CopyDesignFile)
			design.Post("/file/upload", manageController.UploadDesignFile)
			design.Post("/file/delete", manageController.DeleteDesignFile)
			design.Get("/docs", manageController.GetDesignDocs)
			design.Get("/file/templates", manageController.GetDesignTemplateFiles)
		}

		plugin := manage.Party("/plugin", middleware.ParseAdminToken, middleware.AdminPermission)
		{
			plugin.Get("/push", manageController.PluginPush)
			plugin.Post("/push", manageController.PluginPushForm)
			plugin.Get("/push/logs", manageController.PluginPushLogList)

			plugin.Get("/robots", manageController.PluginRobots)
			plugin.Post("/robots", manageController.PluginRobotsForm)

			plugin.Get("/sitemap", manageController.PluginSitemap)
			plugin.Post("/sitemap", manageController.PluginSitemapForm)
			plugin.Post("/sitemap/build", manageController.PluginSitemapBuild)

			plugin.Get("/rewrite", manageController.PluginRewrite)
			plugin.Post("/rewrite", manageController.PluginRewriteForm)

			plugin.Get("/storage", manageController.PluginStorageConfig)
			plugin.Post("/storage", manageController.PluginStorageConfigForm)

			link := plugin.Party("/link")
			{
				link.Get("/list", manageController.PluginLinkList)
				link.Post("/detail", manageController.PluginLinkDetailForm)
				link.Post("/delete", manageController.PluginLinkDelete)
				link.Post("/check", manageController.PluginLinkCheck)
			}

			comment := plugin.Party("/comment")
			{
				comment.Get("/list", manageController.PluginCommentList)
				comment.Get("/detail", manageController.PluginCommentDetail)
				comment.Post("/detail", manageController.PluginCommentDetailForm)
				comment.Post("/delete", manageController.PluginCommentDelete)
				comment.Post("/check", manageController.PluginCommentCheck)
			}

			anchor := plugin.Party("/anchor")
			{
				anchor.Get("/list", manageController.PluginAnchorList)
				anchor.Get("/detail", manageController.PluginAnchorDetail)
				anchor.Post("/detail", manageController.PluginAnchorDetailForm)
				anchor.Post("/replace", manageController.PluginAnchorReplace)
				anchor.Post("/delete", manageController.PluginAnchorDelete)
				anchor.Post("/export", manageController.PluginAnchorExport)
				anchor.Post("/import", manageController.PluginAnchorImport)
				anchor.Get("/setting", manageController.PluginAnchorSetting)
				anchor.Post("/setting", manageController.PluginAnchorSettingForm)
			}

			guestbook := plugin.Party("/guestbook")
			{
				guestbook.Get("/list", manageController.PluginGuestbookList)
				guestbook.Post("/delete", manageController.PluginGuestbookDelete)
				guestbook.Post("/export", manageController.PluginGuestbookExport)
				guestbook.Get("/setting", manageController.PluginGuestbookSetting)
				guestbook.Post("/setting", manageController.PluginGuestbookSettingForm)
			}

			keyword := plugin.Party("/keyword")
			{
				keyword.Get("/setting", manageController.PluginKeywordSetting)
				keyword.Post("/setting", manageController.PluginSaveKeywordSetting)
				keyword.Get("/list", manageController.PluginKeywordList)
				keyword.Post("/detail", manageController.PluginKeywordDetailForm)
				keyword.Post("/delete", manageController.PluginKeywordDelete)
				keyword.Post("/export", manageController.PluginKeywordExport)
				keyword.Post("/import", manageController.PluginKeywordImport)
			}

			fileUpload := plugin.Party("/fileupload")
			{
				fileUpload.Get("/list", manageController.PluginFileUploadList)
				fileUpload.Post("/upload", manageController.PluginFileUploadUpload)
				fileUpload.Post("/delete", manageController.PluginFileUploadDelete)
			}

			material := plugin.Party("/material")
			{
				material.Post("/convert/file", manageController.ConvertFileToUtf8)
				material.Get("/list", manageController.PluginMaterialList)
				material.Post("/detail", manageController.PluginMaterialDetailForm)
				material.Post("/import", manageController.PluginMaterialImport)
				material.Post("/delete", manageController.PluginMaterialDelete)

				material.Get("/category/list", manageController.PluginMaterialCategoryList)
				material.Post("/category/detail", manageController.PluginMaterialCategoryDetailForm)
				material.Post("/category/delete", manageController.PluginMaterialCategoryDelete)
			}

			sendmail := plugin.Party("/sendmail")
			{
				sendmail.Get("/list", manageController.PluginSendmailList)
				sendmail.Get("/setting", manageController.PluginSendmailSetting)
				sendmail.Post("/setting", manageController.PluginSendmailSettingForm)
				sendmail.Post("/test", manageController.PluginSendmailTest)
			}

			importApi := plugin.Party("/import")
			{
				importApi.Get("/api", manageController.PluginImportApi)
				importApi.Post("/token", manageController.PluginUpdateApiToken)
			}

			tag := plugin.Party("/tag")
			{
				tag.Get("/list", manageController.PluginTagList)
				tag.Get("/detail", manageController.PluginTagDetail)
				tag.Post("/detail", manageController.PluginTagDetailForm)
				tag.Post("/delete", manageController.PluginTagDelete)
			}

			redirect := plugin.Party("/redirect")
			{
				redirect.Get("/list", manageController.PluginRedirectList)
				redirect.Post("/detail", manageController.PluginRedirectDetailForm)
				redirect.Post("/delete", manageController.PluginRedirectDelete)
				redirect.Post("/import", manageController.PluginRedirectImport)
			}

			transfer := plugin.Party("/transfer")
			{
				transfer.Get("/task", manageController.GetTransferTask)
				transfer.Post("/download", manageController.DownloadClientFile)
				transfer.Post("/create", manageController.CreateTransferTask)
				transfer.Post("/start", manageController.TransferWebData)
			}

			user := plugin.Party("/user")
			{
				user.Get("/fields", manageController.PluginUserFieldsSetting)
				user.Post("/fields", manageController.PluginUserFieldsSettingForm)
				user.Get("/list", manageController.PluginUserList)
				user.Get("/detail", manageController.PluginUserDetail)
				user.Post("/detail", manageController.PluginUserDetailForm)
				user.Post("/delete", manageController.PluginUserDelete)
				user.Get("/group/list", manageController.PluginUserGroupList)
				user.Get("/group/detail", manageController.PluginUserGroupDetail)
				user.Post("/group/detail", manageController.PluginUserGroupDetailForm)
				user.Post("/group/delete", manageController.PluginUserGroupDelete)
			}

			weapp := plugin.Party("/weapp")
			{
				weapp.Get("/config", manageController.PluginWeappConfig)
				weapp.Post("/config", manageController.PluginWeappConfigForm)
			}

			wechat := plugin.Party("/wechat")
			{
				wechat.Get("/config", manageController.PluginWechatConfig)
				wechat.Post("/config", manageController.PluginWechatConfigForm)
				wechat.Get("/message/list", manageController.PluginWechatMessages)
				wechat.Post("/message/delete", manageController.PluginWechatMessageDelete)
				wechat.Post("/message/reply", manageController.PluginWechatMessageReply)
				wechat.Get("/reply/rule/list", manageController.PluginWechatReplyRules)
				wechat.Post("/reply/rule/delete", manageController.PluginWechatReplyRuleDelete)
				wechat.Post("/reply/rule/save", manageController.PluginWechatReplyRuleForm)
				wechat.Get("/menu/list", manageController.PluginWechatMenus)
				wechat.Post("/menu/save", manageController.PluginWechatMenuSave)
				wechat.Post("/menu/delete", manageController.PluginWechatMenuDelete)
				wechat.Post("/menu/sync", manageController.PluginWechatMenuSync)
			}

			retailer := plugin.Party("/retailer")
			{
				retailer.Get("/list", manageController.PluginGetRetailers)
				retailer.Get("/config", manageController.PluginRetailerConfig)
				retailer.Post("/config", manageController.PluginRetailerConfigForm)
				retailer.Post("/realname", manageController.PluginRetailerSetRealName)
				retailer.Post("/apply", manageController.PluginRetailerApply)
			}

			pay := plugin.Party("/pay")
			{
				pay.Get("/config", manageController.PluginPayConfig)
				pay.Post("/config", manageController.PluginPayConfigForm)
				pay.Post("/upload", manageController.PluginPayUploadFile)
			}

			order := plugin.Party("/order")
			{
				order.Get("/config", manageController.PluginOrderConfig)
				order.Post("/config", manageController.PluginOrderConfigForm)
				order.Get("/list", manageController.PluginOrderList)
				order.Get("/detail", manageController.PluginOrderDetail)
				order.Post("/deliver", manageController.PluginOrderSetDeliver)
				order.Post("/finished", manageController.PluginOrderSetFinished)
				order.Post("/canceled", manageController.PluginOrderSetCanceled)
				order.Post("/refund", manageController.PluginOrderSetRefund)
				order.Post("/refund/apply", manageController.PluginOrderApplyRefund)
				order.Post("/export", manageController.PluginOrderExport)
			}

			withdraw := plugin.Party("/withdraw")
			{
				withdraw.Get("/list", manageController.PluginWithdrawList)
				withdraw.Get("/detail", manageController.PluginWithdrawDetail)
				withdraw.Post("/approval", manageController.PluginWithdrawSetApproval)
				withdraw.Post("/finished", manageController.PluginWithdrawSetFinished)
			}

			commission := plugin.Party("/commission")
			{
				commission.Get("/list", manageController.PluginCommissionList)
				commission.Get("/detail", manageController.PluginCommissionDetail)
			}

			finance := plugin.Party("/finance")
			{
				finance.Get("/list", manageController.PluginFinanceList)
				finance.Get("/detail", manageController.PluginFinanceDetail)
			}

			// 全文索引
			fulltext := plugin.Party("/fulltext")
			{
				fulltext.Get("/config", manageController.PluginFulltextConfig)
				fulltext.Post("/config", manageController.PluginFulltextConfigForm)
			}

			backup := plugin.Party("/backup")
			{
				backup.Get("/list", manageController.PluginBackupList)
				backup.Post("/dump", manageController.PluginBackupDump)
				backup.Post("/restore", manageController.PluginBackupRestore)
				backup.Post("/delete", manageController.PluginBackupDelete)
				backup.Post("/export", manageController.PluginBackupExport)
				backup.Post("/import", manageController.PluginBackupImport)
			}
		}
	}
}
