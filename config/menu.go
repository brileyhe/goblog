package config

type MenuGroup struct {
	Key   string  `json:"key"`
	Name  string  `json:"name"`
	Menus []*Menu `json:"menus"`
}

type Menu struct {
	GroupKey string `json:"group_key"`
	Name     string `json:"name"`
	Path     string `json:"path"` // 前端路由
	Backend  string `json:"-"`    // 后端路由
}

var DefaultMenuGroups = []*MenuGroup{
	{
		Key:  "setting",
		Name: "后台设置",
		Menus: []*Menu{
			{
				Path:     "/setting/system",
				GroupKey: "setting",
				Name:     "全局设置",
				Backend:  "/setting/system",
			},
			{
				Path:     "/setting/content",
				GroupKey: "setting",
				Name:     "内容设置",
				Backend:  "/setting/content",
			},
			{
				Path:     "/setting/safe",
				GroupKey: "setting",
				Name:     "内容安全设置",
				Backend:  "/setting/safe",
			},
			{
				Path:     "/setting/contact",
				GroupKey: "setting",
				Name:     "联系方式设置",
				Backend:  "/setting/contact",
			},
			{
				Path:     "/setting/tdk",
				GroupKey: "setting",
				Name:     "首页TDK设置",
				Backend:  "/setting/index",
			},
			{
				Path:     "/setting/nav",
				GroupKey: "setting",
				Name:     "导航设置",
				Backend:  "/setting/nav",
			},
		},
	},
	{
		Key:  "archive",
		Name: "内容管理",
		Menus: []*Menu{
			{
				Path:     "/archive/list",
				GroupKey: "archive",
				Name:     "文档列表",
				Backend:  "/archive/list",
			},
			{
				Path:     "/archive/recycle",
				GroupKey: "archive",
				Name:     "文档回收站",
				Backend:  "/archive/list",
			},
			{
				Path:     "/archive/detail",
				GroupKey: "archive",
				Name:     "文档编辑",
				Backend:  "/archive/detail",
			},
			{
				Path:     "/archive/category",
				GroupKey: "archive",
				Name:     "文档分类",
			},
			{
				Path:     "/archive/tag",
				GroupKey: "archive",
				Name:     "文档标签",
				Backend:  "/plugin/tag",
			},
			{
				Path:     "/archive/module",
				GroupKey: "archive",
				Name:     "内容模型",
				Backend:  "/module",
			},
		},
	},
	{
		Key:  "content",
		Name: "页面资源",
		Menus: []*Menu{
			{
				Path:     "/content/page",
				GroupKey: "content",
				Name:     "单页面管理",
				Backend:  "/category/list",
			},
			{
				Path:     "/content/attachment",
				GroupKey: "content",
				Name:     "图片资源管理",
				Backend:  "/attachment",
			},
		},
	},
	{
		Key:  "plugin",
		Name: "功能资源",
		Menus: []*Menu{
			{
				Path:     "/plugin/index",
				GroupKey: "plugin",
				Name:     "功能列表",
				Backend:  "/plugin/index",
			},
			{
				Path:     "/plugin/rewrite",
				GroupKey: "plugin",
				Name:     "伪静态规则管理",
				Backend:  "/plugin/rewrite",
			},
			{
				Path:     "/plugin/push",
				GroupKey: "plugin",
				Name:     "链接推送管理",
				Backend:  "/plugin/push",
			},
			{
				Path:     "/plugin/sitemap",
				GroupKey: "plugin",
				Name:     "Sitemap管理",
				Backend:  "/plugin/sitemap",
			},
			{
				Path:     "/plugin/robots",
				GroupKey: "plugin",
				Name:     "Robots管理",
				Backend:  "/plugin/robots",
			},
			{
				Path:     "/plugin/friendlink",
				GroupKey: "plugin",
				Name:     "友情链接管理",
				Backend:  "/plugin/link",
			},
			{
				Path:     "/plugin/comment",
				GroupKey: "plugin",
				Name:     "内容评论管理",
				Backend:  "/plugin/comment",
			},
			{
				Path:     "/plugin/guestbook",
				GroupKey: "plugin",
				Name:     "网站留言管理",
				Backend:  "/plugin/guestbook",
			},
			{
				Path:     "/plugin/keyword",
				GroupKey: "plugin",
				Name:     "关键词库管理",
				Backend:  "/plugin/keyword",
			},
			{
				Path:     "/plugin/material",
				GroupKey: "plugin",
				Name:     "内容素材管理",
				Backend:  "/plugin/material",
			},
			{
				Path:     "/plugin/sendmail",
				GroupKey: "plugin",
				Name:     "邮件提醒",
				Backend:  "/plugin/sendmail",
			},
			{
				Path:     "/plugin/collector",
				GroupKey: "plugin",
				Name:     "内容采集管理",
				Backend:  "/collector",
			},
			{
				Path:     "/plugin/importapi",
				GroupKey: "plugin",
				Name:     "内容导入接口",
				Backend:  "/plugin/import",
			},
			{
				Path:     "/plugin/redirect",
				GroupKey: "plugin",
				Name:     "301跳转管理",
				Backend:  "/plugin/redirect",
			},
			{
				Path:     "/plugin/transfer",
				GroupKey: "plugin",
				Name:     "网站内容迁移",
				Backend:  "/plugin/transfer",
			},
			{
				Path:     "/plugin/storage",
				GroupKey: "plugin",
				Name:     "资源存储配置",
				Backend:  "/plugin/storage",
			},
			{
				Path:     "/plugin/user",
				GroupKey: "plugin",
				Name:     "用户管理",
				Backend:  "/plugin/user",
			},
			{
				Path:     "/plugin/user/group",
				GroupKey: "plugin",
				Name:     "用户组管理",
				Backend:  "/plugin/user/group",
			},
			{
				Path:     "/plugin/weapp",
				GroupKey: "plugin",
				Name:     "小程序配置",
				Backend:  "/plugin/weapp",
			},
			{
				Path:     "/plugin/pay",
				GroupKey: "plugin",
				Name:     "支付配置",
				Backend:  "/plugin/pay",
			},
			{
				Path:     "/plugin/order",
				GroupKey: "plugin",
				Name:     "订单管理",
				Backend:  "/plugin/order",
			},
			{
				Path:     "/plugin/withdraw",
				GroupKey: "plugin",
				Name:     "提现管理",
				Backend:  "/plugin/withdraw",
			},
		},
	},
	{
		Key:  "design",
		Name: "模板设计",
		Menus: []*Menu{
			{
				Path:     "/design/index",
				GroupKey: "design",
				Name:     "我的模板",
				Backend:  "/design/list",
			},
			{
				Path:     "/design/editor",
				GroupKey: "design",
				Name:     "修改代码",
				Backend:  "/design/file",
			},
			{
				Path:     "/design/detail",
				GroupKey: "design",
				Name:     "模板管理",
				Backend:  "/design/info",
			},
			{
				Path:     "/design/doc",
				GroupKey: "design",
				Name:     "开发文档",
				Backend:  "/design/docs",
			},
			{
				Path:     "/design/market",
				GroupKey: "design",
				Name:     "设计市场",
				Backend:  "/design/market",
			},
		},
	},
	{
		Key:  "statistic",
		Name: "数据统计",
		Menus: []*Menu{
			{
				Path:     "/statistic/spider",
				GroupKey: "statistic",
				Name:     "蜘蛛统计",
			},
			{
				Path:     "/statistic/traffic",
				GroupKey: "statistic",
				Name:     "流量统计",
			},
			{
				Path:     "/statistic/detail",
				GroupKey: "statistic",
				Name:     "访问详细记录",
			},
			{
				Path:     "/statistic/includes",
				GroupKey: "statistic",
				Name:     "收录统计",
			},
			{
				Path:     "/statistic/include/detail",
				GroupKey: "statistic",
				Name:     "收录详细记录",
			},
		},
	},
	{
		Key:  "account",
		Name: "管理员",
		Menus: []*Menu{
			{
				Path:     "/account/list",
				GroupKey: "account",
				Name:     "管理员列表",
				Backend:  "/admin/list",
			},
			{
				Path:     "/account/index",
				GroupKey: "account",
				Name:     "管理员信息",
			},
			{
				Path:     "/account/group/list",
				GroupKey: "account",
				Name:     "管理员分组列表",
				Backend:  "/admin/group/list",
			},
			{
				Path:     "/account/group/detail",
				GroupKey: "account",
				Name:     "管理员分组信息",
				Backend:  "/admin/group/detail",
			},
			{
				Path:     "/account/logs/login",
				GroupKey: "account",
				Name:     "登录记录",
				Backend:  "/admin/logs/login",
			},
			{
				Path:     "/account/logs/action",
				GroupKey: "account",
				Name:     "操作记录",
				Backend:  "/admin/logs/action",
			},
		},
	},
	{
		Key:  "tool",
		Name: "系统功能",
		Menus: []*Menu{
			{
				Path:     "/tool/upgrade",
				GroupKey: "tool",
				Name:     "操作记录",
				Backend:  "/version/upgrade",
			},
			{
				Path:     "/tool/cache",
				GroupKey: "tool",
				Name:     "更新缓存",
				Backend:  "/setting/cache",
			},
		},
	},
}
