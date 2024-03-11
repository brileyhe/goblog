package provider

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/request"
	"kandaoni.com/anqicms/response"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (w *Website) GetCategories(ops func(tx *gorm.DB) *gorm.DB, parentId uint) ([]*model.Category, error) {
	var categories []*model.Category
	err := ops(w.DB).Find(&categories).Error
	if err != nil {
		return nil, err
	}
	for i := range categories {
		categories[i].GetThumb(w.PluginStorage.StorageUrl, w.Content.DefaultThumb)
		categories[i].Link = w.GetUrl("category", categories[i], 0)
	}
	categoryTree := NewCategoryTree(categories)
	categories = categoryTree.GetTree(parentId, "")

	return categories, nil
}

func (w *Website) GetCategoryByTitle(title string) (*model.Category, error) {
	return w.GetCategoryByFunc(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("`title` = ?", title)
	})
}

func (w *Website) GetCategoryById(id uint) (*model.Category, error) {
	return w.GetCategoryByFunc(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("`id` = ?", id)
	})
}

func (w *Website) GetCategoryByUrlToken(urlToken string) (*model.Category, error) {
	if urlToken == "" {
		return nil, errors.New("empty token")
	}
	return w.GetCategoryByFunc(func(tx *gorm.DB) *gorm.DB {
		return tx.Where("`url_token` = ?", urlToken)
	})
}

func (w *Website) GetCategoryByFunc(ops func(tx *gorm.DB) *gorm.DB) (*model.Category, error) {
	var category model.Category
	err := ops(w.DB).Take(&category).Error
	if err != nil {
		return nil, err
	}
	category.GetThumb(w.PluginStorage.StorageUrl, w.Content.DefaultThumb)
	category.Link = w.GetUrl("category", &category, 0)

	return &category, nil
}

func (w *Website) SaveCategory(req *request.Category) (category *model.Category, err error) {
	newPost := false
	if req.Id > 0 {
		category, err = w.GetCategoryById(req.Id)
		if err != nil {
			return nil, err
		}
	} else {
		category = &model.Category{
			Status: 1,
		}
		newPost = true
	}
	category.Title = req.Title
	category.SeoTitle = req.SeoTitle
	category.Keywords = req.Keywords
	category.Description = req.Description
	category.Type = req.Type
	category.ModuleId = req.ModuleId
	category.ParentId = req.ParentId
	category.Sort = req.Sort
	category.Status = req.Status
	category.Template = req.Template
	category.DetailTemplate = req.DetailTemplate
	category.IsInherit = req.IsInherit
	category.Images = req.Images
	category.Logo = req.Logo
	for i, v := range category.Images {
		category.Images[i] = strings.TrimPrefix(v, w.PluginStorage.StorageUrl)
	}
	if category.Logo != "" {
		category.Logo = strings.TrimPrefix(category.Logo, w.PluginStorage.StorageUrl)
	}
	//增加判断上级，强制类型与上级同步
	if category.ParentId > 0 {
		parent, err := w.GetCategoryById(category.ParentId)
		if err == nil {
			category.Type = parent.Type
			category.ModuleId = parent.ModuleId
		}
	}
	// 判断重复
	req.UrlToken = library.ParseUrlToken(req.UrlToken)
	if req.UrlToken == "" {
		req.UrlToken = library.GetPinyin(req.Title, w.Content.UrlTokenType == config.UrlTokenTypeSort)
	}
	category.UrlToken = w.VerifyCategoryUrlToken(req.UrlToken, category.Id)
	if category.ModuleId == 0 && category.Type == config.CategoryTypeArchive {
		modules := w.GetCacheModules()
		if len(modules) > 0 {
			category.ModuleId = modules[0].Id
		}
	}
	if category.Type == config.CategoryTypePage {
		category.ModuleId = 0
	}
	// 将单个&nbsp;替换为空格
	req.Content = library.ReplaceSingleSpace(req.Content)
	req.Content = strings.ReplaceAll(req.Content, w.System.BaseUrl, "")
	baseHost := ""
	urls, err := url.Parse(w.System.BaseUrl)
	if err == nil {
		baseHost = urls.Host
	}
	autoAddImage := false
	//提取描述
	if category.Description == "" {
		tmpContent := req.Content
		if w.Content.Editor == "markdown" {
			tmpContent = library.MarkdownToHTML(tmpContent)
		}
		category.Description = library.ParseDescription(strings.ReplaceAll(CleanTagsAndSpaces(tmpContent), "\n", " "))
	}
	//提取缩略图
	if len(category.Logo) == 0 {
		re, _ := regexp.Compile(`(?i)<img.*?src="(.+?)".*?>`)
		match := re.FindStringSubmatch(req.Content)
		if len(match) > 1 {
			//提取缩略图
			category.Logo = match[1]
			autoAddImage = true
		} else {
			// 匹配Markdown ![新的图片](http://xxx/xxx.webp)
			re, _ = regexp.Compile(`!\[([^]]*)\]\(([^)]+)\)`)
			match = re.FindStringSubmatch(req.Content)
			if len(match) > 2 {
				category.Logo = match[2]
				autoAddImage = true
			}
		}
	}
	// 过滤外链
	if w.Content.FilterOutlink == 1 {
		re, _ := regexp.Compile(`(?i)<a.*?href="(.+?)".*?>(.*?)</a>`)
		req.Content = re.ReplaceAllStringFunc(req.Content, func(s string) string {
			match := re.FindStringSubmatch(s)
			if len(match) < 3 {
				return s
			}
			aUrl, err2 := url.Parse(match[1])
			if err2 == nil {
				if aUrl.Host != "" && aUrl.Host != baseHost {
					//过滤外链
					return match[2]
				}
			}
			return s
		})
		// 匹配Markdown [link](url)
		// 由于不支持零宽断言，因此匹配所有
		re, _ = regexp.Compile(`!?\[([^]]*)\]\(([^)]+)\)`)
		req.Content = re.ReplaceAllStringFunc(req.Content, func(s string) string {
			// 过滤掉 ! 开头的
			if strings.HasPrefix(s, "!") {
				return s
			}
			match := re.FindStringSubmatch(s)
			if len(match) < 3 {
				return s
			}
			aUrl, err2 := url.Parse(match[2])
			if err2 == nil {
				if aUrl.Host != "" && aUrl.Host != baseHost {
					//过滤外链
					return match[1]
				}
			}
			return s
		})
	}
	category.Content = req.Content

	err = category.Save(w.DB)
	if err != nil {
		return
	}
	//检查有多少个material
	var materialIds []uint
	re, _ := regexp.Compile(`(?i)<div.*?data-material="(\d+)".*?>`)
	matches := re.FindAllStringSubmatch(req.Content, -1)
	if len(matches) > 0 {
		for _, match := range matches {
			//记录material
			materialId, _ := strconv.Atoi(match[1])
			if materialId > 0 {
				materialIds = append(materialIds, uint(materialId))
			}
		}
	}
	go w.LogMaterialData(materialIds, "category", category.Id)
	// 自动提取远程图片改成保存后处理
	if w.Content.RemoteDownload == 1 {
		hasChangeImg := false
		re, _ = regexp.Compile(`(?i)<img.*?src="(.+?)".*?>`)
		category.Content = re.ReplaceAllStringFunc(category.Content, func(s string) string {
			match := re.FindStringSubmatch(s)
			if len(match) < 2 {
				return s
			}
			imgUrl, err2 := url.Parse(match[1])
			if err2 == nil {
				if imgUrl.Host != "" && imgUrl.Host != baseHost && !strings.HasPrefix(match[1], w.PluginStorage.StorageUrl) {
					//外链
					attachment, err2 := w.DownloadRemoteImage(match[1], "")
					if err2 == nil {
						// 下载完成
						hasChangeImg = true
						s = strings.Replace(s, match[1], attachment.Logo, 1)
					}
				}
			}
			return s
		})
		// 匹配Markdown ![新的图片](http://xxx/xxx.webp)
		re, _ = regexp.Compile(`!\[([^]]*)\]\(([^)]+)\)`)
		category.Content = re.ReplaceAllStringFunc(category.Content, func(s string) string {
			match := re.FindStringSubmatch(s)
			if len(match) < 3 {
				return s
			}
			imgUrl, err2 := url.Parse(match[2])
			if err2 == nil {
				if imgUrl.Host != "" && imgUrl.Host != baseHost && !strings.HasPrefix(match[2], w.PluginStorage.StorageUrl) {
					//外链
					attachment, err2 := w.DownloadRemoteImage(match[2], "")
					if err2 == nil {
						// 下载完成
						hasChangeImg = true
						s = strings.Replace(s, match[2], attachment.Logo, 1)
					}
				}
			}
			return s
		})
		if hasChangeImg {
			w.DB.Model(category).UpdateColumn("content", category.Content)
			// 更新data
			if autoAddImage {
				//提取缩略图
				re, _ = regexp.Compile(`(?i)<img.*?src="(.+?)".*?>`)
				match := re.FindStringSubmatch(req.Content)
				if len(match) > 1 {
					category.Logo = match[1]
				} else {
					// 匹配Markdown ![新的图片](http://xxx/xxx.webp)
					re, _ = regexp.Compile(`!\[([^]]*)\]\(([^)]+)\)`)
					match = re.FindStringSubmatch(req.Content)
					if len(match) > 2 {
						category.Logo = match[2]
					}
				}
				w.DB.Model(category).UpdateColumn("logo", category.Logo)
			}
		}
	}
	// 如果隐藏的分类有下级，则下级也隐藏
	if category.Status == config.ContentStatusDraft {
		w.DB.Model(&model.Category{}).Where("`parent_id` = ?", category.Id).UpdateColumn("status", config.ContentStatusDraft)
	} else if category.Status == config.ContentStatusOK && category.ParentId > 0 {
		w.DB.Model(&model.Category{}).Where("`id` = ?", category.ParentId).UpdateColumn("status", config.ContentStatusOK)
	}

	if newPost && category.Status == config.ContentStatusOK {
		link := w.GetUrl("category", category, 0)
		go func() {
			w.PushArchive(link)
			if w.PluginSitemap.AutoBuild == 1 {
				_ = w.AddonSitemap("category", link, time.Unix(category.UpdatedTime, 0).Format("2006-01-02"))
			}
		}()
	}
	category.GetThumb(w.PluginStorage.StorageUrl, w.Content.DefaultThumb)
	w.DeleteCacheCategories()
	w.DeleteCacheIndex()

	return
}

// GetCategoryTemplate 获取分类模板，如果检测到不继承，则停止获取
func (w *Website) GetCategoryTemplate(category *model.Category) *response.CategoryTemplate {
	if category == nil {
		return nil
	}

	if category.Template != "" || category.DetailTemplate != "" {
		return &response.CategoryTemplate{
			Template:       category.Template,
			DetailTemplate: category.DetailTemplate,
		}
	}

	//查找上级
	if category.ParentId > 0 {
		parent := w.GetCategoryFromCache(category.ParentId)
		if parent != nil {
			// 如果上级存在模板，并且选择不继承，从这里阻止
			if parent.Template != "" && parent.IsInherit == 0 {
				return nil
			}
		}
		return w.GetCategoryTemplate(parent)
	}

	//不存在，则返回空
	return nil
}

func (w *Website) DeleteCacheCategories() {
	w.Cache.Delete("categories")
}

func (w *Website) GetCacheCategories() []*model.Category {
	if w.DB == nil {
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	var categories []*model.Category

	err := w.Cache.Get("categories", &categories)

	if err == nil {
		return categories
	}

	w.DB.Model(model.Category{}).Order("sort asc").Find(&categories)
	for i := range categories {
		categories[i].GetThumb(w.PluginStorage.StorageUrl, w.Content.DefaultThumb)
	}
	categoryTree := NewCategoryTree(categories)
	categories = categoryTree.GetTree(0, "")

	_ = w.Cache.Set("categories", categories, 0)

	return categories
}

// GetSubCategoryIds 获取分类的子分类
func (w *Website) GetSubCategoryIds(categoryId uint, categories []*model.Category) []uint {
	var subIds []uint
	if categories == nil {
		categories = w.GetCacheCategories()
	}

	for i := range categories {
		if categories[i].Status != config.ContentStatusOK {
			continue
		}
		if categories[i].ParentId == categoryId {
			subIds = append(subIds, categories[i].Id)
			subIds = append(subIds, w.GetSubCategoryIds(categories[i].Id, categories)...)
		}
	}

	return subIds
}

func (w *Website) GetCategoryFromCache(categoryId uint) *model.Category {
	if categoryId == 0 {
		return nil
	}
	categories := w.GetCacheCategories()
	for i := range categories {
		if categories[i].Id == categoryId {
			return categories[i]
		}
	}

	return nil
}

func (w *Website) GetCategoryFromCacheByToken(urlToken string) *model.Category {
	categories := w.GetCacheCategories()
	for i := range categories {
		if categories[i].UrlToken == urlToken {
			return categories[i]
		}
	}

	return nil
}

func (w *Website) GetCategoriesFromCache(moduleId, parentId uint, pageType int, all bool) []*model.Category {
	categories := w.GetCacheCategories()
	var tmpCategories = make([]*model.Category, 0, len(categories))
	for i := range categories {
		if categories[i].Status != config.ContentStatusOK {
			// 跳过隐藏的分类
			continue
		}
		if categories[i].Type != uint(pageType) {
			continue
		}
		if moduleId > 0 && pageType != config.CategoryTypePage {
			if categories[i].ModuleId != moduleId {
				continue
			}
		}
		if all || categories[i].ParentId == parentId {
			tmpCategories = append(tmpCategories, categories[i])
		}
	}

	return tmpCategories
}

func (w *Website) VerifyCategoryUrlToken(urlToken string, id uint) string {
	index := 0
	// 防止超出长度
	if len(urlToken) > 150 {
		urlToken = urlToken[:150]
	}
	urlToken = strings.ToLower(urlToken)
	for {
		tmpToken := urlToken
		if index > 0 {
			tmpToken = fmt.Sprintf("%s-%d", urlToken, index)
		}
		// 判断分类
		tmpCat, err := w.GetCategoryByUrlToken(tmpToken)
		if err == nil && tmpCat.Id != id {
			index++
			continue
		}
		// 判断archive
		_, err = w.GetArchiveByUrlToken(tmpToken)
		if err == nil {
			index++
			continue
		}
		urlToken = tmpToken
		break
	}

	return urlToken
}
