package tags

import (
	"fmt"
	"github.com/flosch/pongo2/v4"
	"kandaoni.com/anqicms/config"
	"kandaoni.com/anqicms/dao"
	"kandaoni.com/anqicms/provider"
)

type tagPageListNode struct {
	name    string
	args    map[string]pongo2.IEvaluator
	wrapper *pongo2.NodeWrapper
}

func (node *tagPageListNode) Execute(ctx *pongo2.ExecutionContext, writer pongo2.TemplateWriter) *pongo2.Error {
	if dao.DB == nil {
		return nil
	}
	pageList := provider.GetCategoriesFromCache(0, 0, config.CategoryTypePage)
	for i := range pageList {
		pageList[i].Link = provider.GetUrl("page", pageList[i], 0)
	}

	ctx.Private[node.name] = pageList
	//execute
	node.wrapper.Execute(ctx, writer)

	return nil
}

func TagPageListParser(doc *pongo2.Parser, start *pongo2.Token, arguments *pongo2.Parser) (pongo2.INodeTag, *pongo2.Error) {
	tagNode := &tagPageListNode{
		args: make(map[string]pongo2.IEvaluator),
	}

	nameToken := arguments.MatchType(pongo2.TokenIdentifier)
	if nameToken == nil {
		return nil, arguments.Error("pageList-tag needs a accept name.", nil)
	}

	tagNode.name = nameToken.Val

	// After having parsed the name we're gonna parse the with options
	args, err := parseWith(arguments)
	if err != nil {
		return nil, err
	}
	tagNode.args = args

	for arguments.Remaining() > 0 {
		return nil, arguments.Error("Malformed pageList-tag arguments.", nil)
	}
	wrapper, endtagargs, err := doc.WrapUntilTag("endpageList")
	if err != nil {
		return nil, err
	}
	if endtagargs.Remaining() > 0 {
		endtagnameToken := endtagargs.MatchType(pongo2.TokenIdentifier)
		if endtagnameToken != nil {
			if endtagnameToken.Val != nameToken.Val {
				return nil, endtagargs.Error(fmt.Sprintf("Name for 'endpageList' must equal to 'pageList'-tag's name ('%s' != '%s').",
					nameToken.Val, endtagnameToken.Val), nil)
			}
		}

		if endtagnameToken == nil || endtagargs.Remaining() > 0 {
			return nil, endtagargs.Error("Either no or only one argument (identifier) allowed for 'endpageList'.", nil)
		}
	}
	tagNode.wrapper = wrapper

	return tagNode, nil
}
