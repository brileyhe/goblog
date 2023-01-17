package tags

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/dao"
	"kandaoni.com/anqicms/library"
	"kandaoni.com/anqicms/model"
	"kandaoni.com/anqicms/provider"
	"reflect"
	"time"
)

type tagArchiveDetailNode struct {
	args map[string]pongo2.IEvaluator
	name string
}

func (node *tagArchiveDetailNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	if dao.DB == nil {
		return nil
	}
	args, err := parseArgs(node.args, ctx)
	if err != nil {
		return err
	}
	id := uint(0)

	fieldName := ""
	if args["name"] != nil {
		fieldName = args["name"].String()
		fieldName = library.Case2Camel(fieldName)
	}

	format := "2006-01-02"
	if args["format"] != nil {
		format = args["format"].String()
	}

	archiveDetail, _ := ctx.Public["archive"].(*model.Archive)

	if args["id"] != nil {
		id = uint(args["id"].Integer())
		archiveDetail, _ = provider.GetArchiveById(id)
		// check has Order
		if fieldName == "HasOrdered" && archiveDetail != nil {
			// if read level larger than 0, then need to check permission
			if archiveDetail.Price == 0 && archiveDetail.ReadLevel == 0 {
				archiveDetail.HasOrdered = true
			}
			userInfo, ok := ctx.Public["userInfo"].(*model.User)
			if ok && userInfo.Id > 0 {
				if archiveDetail.UserId == userInfo.Id {
					archiveDetail.HasOrdered = true
				}
				if archiveDetail.Price > 0 {
					archiveDetail.HasOrdered = provider.CheckArchiveHasOrder(userInfo.Id, archiveDetail.Id)
				}
				if archiveDetail.ReadLevel > 0 && !archiveDetail.HasOrdered {
					userGroup, _ := ctx.Public["userGroup"].(*model.UserGroup)
					if userGroup != nil && userGroup.Level >= archiveDetail.ReadLevel {
						archiveDetail.HasOrdered = true
					}
				}
				discount := provider.GetUserDiscount(userInfo.Id, userInfo)
				if discount > 0 {
					archiveDetail.FavorablePrice = archiveDetail.Price * discount / 100
				}
			}
		}
	}

	if archiveDetail != nil {
		v := reflect.ValueOf(*archiveDetail)

		f := v.FieldByName(fieldName)

		content := fmt.Sprintf("%v", f)
		if content == "" && fieldName == "SeoTitle" {
			content = archiveDetail.Title
		}

		if fieldName == "CreatedTime" || fieldName == "UpdatedTime" {
			content = time.Unix(f.Int(), 0).Format(format)
		}
		if fieldName == "Link" {
			// 当是获取链接的时候，再生成
			archiveDetail.Link = provider.GetUrl("archive", archiveDetail, 0)
		}
		if fieldName == "Content" {
			// if read level larger than 0, then need to check permission
			if archiveDetail.ReadLevel > 0 {
				userGroup, _ := ctx.Public["userGroup"].(*model.UserGroup)
				if userGroup == nil || userGroup.Level < archiveDetail.ReadLevel {
					content = fmt.Sprintf(config.Lang("该内容需要用户等级%d以上才能阅读"), archiveDetail.ReadLevel)
				}
			} else {
				// 当读取content 的时候，再查询
				archiveData, err := provider.GetArchiveDataById(archiveDetail.Id)
				if err == nil {
					content = archiveData.Content
				}
			}
		}
		if fieldName == "Images" || fieldName == "Category" {
			content = ""
		}

		var category *model.Category
		if fieldName == "Category" {
			category = provider.GetCategoryFromCache(archiveDetail.CategoryId)
			if category != nil {
				category.Link = provider.GetUrl("category", category, 0)
			}
		}

		// output
		if node.name == "" {
			writer.WriteString(content)
		} else {
			//不是所有都是字符串
			if fieldName == "Images" {
				ctx.Private[node.name] = archiveDetail.Images
			} else if fieldName == "Category" {
				ctx.Private[node.name] = category
			} else {
				ctx.Private[node.name] = content
			}
		}
	}

	return nil
}

func TagArchiveDetailParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagNode := &tagArchiveDetailNode{
		args: make(map[string]pongo2.IEvaluator),
	}

	nameToken := arguments.MatchType(pongo2.TokenIdentifier)
	if nameToken == nil {
		return nil, arguments.Error("System-tag needs a system config name.", nil)
	}

	if nameToken.Val == "with" {
		//with 需要退回
		arguments.ConsumeN(-1)
	} else {
		tagNode.name = nameToken.Val
	}

	args, err := parseWith(arguments)
	if err != nil {
		return nil, err
	}
	tagNode.args = args

	for arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed archiveDetail-tag arguments.", nil)
	}

	return tagNode, nil
}
