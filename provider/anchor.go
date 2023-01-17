package provider

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/dao"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/model"
	"math"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"
)

type AnchorCSV struct {
	Title  string `csv:"title"`
	Link   string `csv:"link"`
	Weight int    `csv:"weight"`
}

func GetAnchorList(keyword string, currentPage, pageSize int) ([]*model.Anchor, int64, error) {
	var anchors []*model.Anchor
	offset := (currentPage - 1) * pageSize
	var total int64

	builder := dao.DB.Model(&model.Anchor{}).Order("id desc")
	if keyword != "" {
		//模糊搜索
		builder = builder.Where("(`title` like ? OR `link` like ?)", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := builder.Count(&total).Limit(pageSize).Offset(offset).Find(&anchors).Error
	if err != nil {
		return nil, 0, err
	}

	return anchors, total, nil
}

func GetAllAnchors() ([]*model.Anchor, error) {
	var anchors []*model.Anchor
	err := dao.DB.Model(&model.Anchor{}).Order("weight desc").Find(&anchors).Error
	if err != nil {
		return nil, err
	}

	return anchors, nil
}

func GetAnchorById(id uint) (*model.Anchor, error) {
	var anchor model.Anchor

	err := dao.DB.Where("`id` = ?", id).First(&anchor).Error
	if err != nil {
		return nil, err
	}

	return &anchor, nil
}

func GetAnchorByTitle(title string) (*model.Anchor, error) {
	var anchor model.Anchor

	err := dao.DB.Where("`title` = ?", title).First(&anchor).Error
	if err != nil {
		return nil, err
	}

	return &anchor, nil
}

func ImportAnchors(file multipart.File, info *multipart.FileHeader) (string, error) {
	buff, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(buff), "\n")
	var total int
	for i, line := range lines {
		line = strings.TrimSpace(line)
		// 格式：title, link, weight
		if i == 0 {
			continue
		}
		values := strings.Split(line, ",")
		if len(values) < 3 {
			continue
		}
		title := strings.TrimSpace(values[0])
		if title == "" {
			continue
		}
		anchor, err := GetAnchorByTitle(title)
		if err != nil {
			//表示不存在
			anchor = &model.Anchor{
				Title:  title,
				Status: 1,
			}
			total++
		}
		anchor.Link = strings.TrimPrefix(values[1], config.JsonData.System.BaseUrl)
		anchor.Weight, _ = strconv.Atoi(values[2])

		anchor.Save(dao.DB)
	}

	return fmt.Sprintf(config.Lang("成功导入了%d个锚文本"), total), nil
}

func DeleteAnchor(anchor *model.Anchor) error {
	err := dao.DB.Delete(anchor).Error
	if err != nil {
		return err
	}

	//清理已经存在的anchor
	go CleanAnchor(anchor.Id)

	return nil
}

func CleanAnchor(anchorId uint) {
	var anchorData []*model.AnchorData
	err := dao.DB.Where("`anchor_id` = ?", anchorId).Find(&anchorData).Error
	if err != nil {
		return
	}

	anchorIdStr := fmt.Sprintf("%d", anchorId)

	for _, data := range anchorData {
		//处理archive
		archiveData, err := GetArchiveDataById(data.ItemId)
		if err != nil {
			continue
		}
		htmlR := strings.NewReader(archiveData.Content)
		doc, err := goquery.NewDocumentFromReader(htmlR)
		if err == nil {
			clean := false
			doc.Find("a,strong").Each(func(i int, s *goquery.Selection) {
				existsId, exists := s.Attr("data-anchor")

				if exists && existsId == anchorIdStr {
					//清理它
					s.Contents().Unwrap()
					clean = true
				}
			})
			//清理完毕，更新
			if clean {
				//更新内容
				archiveData.Content, _ = doc.Find("body").Html()
				dao.DB.Save(archiveData)
			}
		}
		//删除当前item
		dao.DB.Unscoped().Delete(data)
	}
}

func ChangeAnchor(anchor *model.Anchor, changeTitle bool) {
	//如果锚文本更改了名称，需要移除已经生成锚文本
	if changeTitle {
		//清理anchor
		CleanAnchor(anchor.Id)

		//更新替换数量
		anchor.ReplaceCount = 0
		dao.DB.Save(anchor)
		return
	}
	//其他当做更改了连接
	//如果锚文本只更改了连接，则需要重新替换新的连接
	var anchorData []*model.AnchorData
	err := dao.DB.Where("`anchor_id` = ?", anchor.Id).Find(&anchorData).Error
	if err != nil {
		return
	}

	anchorIdStr := fmt.Sprintf("%d", anchor.Id)

	for _, data := range anchorData {
		//处理archive
		archiveData, err := GetArchiveDataById(data.ItemId)
		if err != nil {
			continue
		}
		htmlR := strings.NewReader(archiveData.Content)
		doc, err := goquery.NewDocumentFromReader(htmlR)
		if err == nil {
			update := false
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				existsId, exists := s.Attr("data-anchor")

				if exists && existsId == anchorIdStr {
					//换成新的链接
					s.SetAttr("href", anchor.Link)
					update = true
				}
			})
			//更新完毕，更新
			if update {
				//更新内容
				archiveData.Content, _ = doc.Find("body").Html()
				dao.DB.Save(archiveData)
			}
		}
	}
}

// ReplaceAnchor 单个替换
func ReplaceAnchor(anchor *model.Anchor) {
	//交由下方执行
	if anchor == nil {
		ReplaceAnchors(nil)
	} else {
		ReplaceAnchors([]*model.Anchor{anchor})
	}
}

// ReplaceAnchors 批量替换
func ReplaceAnchors(anchors []*model.Anchor) {
	if len(anchors) == 0 {
		anchors, _ = GetAllAnchors()
		if len(anchors) == 0 {
			//没有关键词，终止执行
			return
		}
	}

	//先遍历文章、产品，添加锚文本
	//每次取100个
	limit := 100
	lastId := uint(0)
	var archives []*model.Archive

	for {
		dao.DB.Where("`id` > ?", lastId).Order("id asc").Limit(limit).Find(&archives)
		if len(archives) == 0 {
			break
		}
		//加下一轮
		lastId = archives[len(archives)-1].Id
		for _, v := range archives {
			//执行替换
			link := GetUrl("archive", v, 0)
			ReplaceContent(anchors, "archive", v.Id, link)
		}
	}
}

func ReplaceContent(anchors []*model.Anchor, itemType string, itemId uint, link string) string {
	link = strings.TrimPrefix(link, config.JsonData.System.BaseUrl)
	if len(anchors) == 0 {
		anchors, _ = GetAllAnchors()
		if len(anchors) == 0 {
			//没有关键词，终止执行
			return ""
		}
	}

	content := ""

	archiveData, err := GetArchiveDataById(itemId)
	if err != nil {
		return ""
	}
	content = archiveData.Content

	//获取纯文本字数
	stripedContent := library.StripTags(content)
	contentLen := len([]rune(stripedContent))
	if config.JsonData.PluginAnchor.AnchorDensity < 20 {
		//默认设置200
		config.JsonData.PluginAnchor.AnchorDensity = 200
	}

	//最大可以替换的数量
	maxAnchorNum := int(math.Ceil(float64(contentLen) / float64(config.JsonData.PluginAnchor.AnchorDensity)))

	type replaceType struct {
		Key   string
		Value string
	}

	existsKeywords := map[string]bool{}
	existsLinks := map[string]bool{}

	var replacedMatch []*replaceType
	numCount := 0
	//所有的a标签计数，并替换掉
	reg, _ := regexp.Compile("(?i)<a[^>]*>(.*?)</a>")
	content = reg.ReplaceAllStringFunc(content, func(s string) string {

		reg := regexp.MustCompile("(?i)<a\\s*[^>]*href=[\"']?([^\"']*)[\"']?[^>]*>(.*?)</a>")
		match := reg.FindStringSubmatch(s)
		if len(match) > 2 {
			existsKeywords[strings.ToLower(match[2])] = true
			existsLinks[strings.ToLower(match[1])] = true
		}

		key := fmt.Sprintf("{$%d}", numCount)
		replacedMatch = append(replacedMatch, &replaceType{
			Key:   key,
			Value: s,
		})
		numCount++

		return key
	})
	//所有的strong标签替换掉
	reg, _ = regexp.Compile("(?i)<strong[^>]*>(.*?)</strong>")
	content = reg.ReplaceAllStringFunc(content, func(s string) string {
		key := fmt.Sprintf("{$%d}", numCount)
		replacedMatch = append(replacedMatch, &replaceType{
			Key:   key,
			Value: s,
		})
		numCount++

		return key
	})
	//过滤所有属性
	reg, _ = regexp.Compile("(?i)</?[a-z0-9]+(\\s+[^>]+)>")
	content = reg.ReplaceAllStringFunc(content, func(s string) string {
		key := fmt.Sprintf("{$%d}", numCount)
		replacedMatch = append(replacedMatch, &replaceType{
			Key:   key,
			Value: s,
		})
		numCount++

		return key
	})

	if len(existsLinks) < maxAnchorNum {
		//开始替换关键词
		for _, anchor := range anchors {
			if anchor.Title == "" {
				continue
			}
			if strings.HasSuffix(anchor.Link, link) {
				//当前url，跳过
				continue
			}
			//已经存在存在的关键词，或者链接，跳过
			if existsKeywords[strings.ToLower(anchor.Title)] || existsLinks[strings.ToLower(anchor.Link)] {
				continue
			}
			//开始替换
			replaceNum := 0
			replacer := strings.NewReplacer("\\", "\\\\", "/", "\\/", "{", "\\{", "}", "\\}", "^", "\\^", "$", "\\$", "*", "\\*", "+", "\\+", "?", "\\?", ".", "\\.", "|", "\\|", "-", "\\-", "[", "\\[", "]", "\\]", "(", "\\(", ")", "\\)")
			matchName := replacer.Replace(anchor.Title)

			reg, _ = regexp.Compile(fmt.Sprintf("(?i)%s", matchName))
			content = reg.ReplaceAllStringFunc(content, func(s string) string {
				replaceHtml := ""
				key := ""
				if replaceNum == 0 {
					//第一条替换为锚文本
					replaceHtml = fmt.Sprintf("<a href=\"%s\" data-anchor=\"%d\">%s</a>", anchor.Link, anchor.Id, s)
					key = fmt.Sprintf("{$%d}", numCount)

					//加入计数
					existsLinks[anchor.Link] = true
					existsKeywords[anchor.Title] = true
				} else {
					//其他则加粗
					replaceHtml = fmt.Sprintf("<strong data-anchor=\"%d\">%s</strong>", anchor.Id, s)
					key = fmt.Sprintf("{$%d}", numCount)
				}
				replaceNum++

				replacedMatch = append(replacedMatch, &replaceType{
					Key:   key,
					Value: replaceHtml,
				})
				numCount++

				return key
			})

			//如果有更新了，则记录
			if replaceNum > 0 {
				//插入记录
				anchorData := &model.AnchorData{
					AnchorId: anchor.Id,
					ItemType: itemType,
					ItemId:   itemId,
				}
				dao.DB.Save(anchorData)
				//更新计数
				var count int64
				dao.DB.Model(&model.AnchorData{}).Where("`anchor_id` = ?", anchor.Id).Count(&count)
				anchor.ReplaceCount = count
				dao.DB.Save(anchor)
			}

			//判断数量是否达到了，达到了就跳出
			if len(existsLinks) >= maxAnchorNum {
				break
			}
		}
	}

	//关键词替换完毕，将原来替换的重新替换回去，需要倒序
	for i := len(replacedMatch) - 1; i >= 0; i-- {
		content = strings.Replace(content, replacedMatch[i].Key, replacedMatch[i].Value, 1)
	}

	if !strings.EqualFold(archiveData.Content, content) {
		//内容有更新，执行更新
		archiveData.Content = content
		dao.DB.Save(archiveData)
	}

	return content
}

func AutoInsertAnchor(archiveId uint, keywords, link string) {
	link = strings.TrimPrefix(link, config.JsonData.System.BaseUrl)
	keywords = strings.ReplaceAll(keywords, "，", ",")
	keywords = strings.ReplaceAll(keywords, " ", ",")
	keywords = strings.ReplaceAll(keywords, "_", ",")

	keywordArr := strings.Split(keywords, ",")
	for _, v := range keywordArr {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		_, err := GetAnchorByTitle(v)
		if err != nil {
			//插入新的
			anchor := &model.Anchor{
				Title:     v,
				Link:      link,
				ArchiveId: archiveId,
				Status:    1,
			}
			dao.DB.Save(anchor)
		}
	}
}
