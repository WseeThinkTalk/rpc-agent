package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/agent/internal/svc"
	"rpc-agent/client/article/article"
)

type SearchArticlesParams struct {
	Keyword  string `json:"keyword" description:"搜索关键词，匹配文章标题和内容"`
	PageSize int32  `json:"page_size" description:"返回数量，1-20"`
}

type SearchArticleItem struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	AuthorName  string `json:"author_name"`
	LikeNum     int64  `json:"like_num"`
	CommentNum  int64  `json:"comment_num"`
	PublishTime string `json:"publish_time"`
}

type SearchArticlesResult struct {
	Articles []SearchArticleItem `json:"articles"`
	Total    int64               `json:"total"`
}

func NewSearchArticlesTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("search_articles",
		"在 ThinkTalk 平台搜索已发布的文章。支持按关键词匹配标题和内容，返回匹配的文章列表及每篇的点赞数、评论数。",
		func(ctx context.Context, params *SearchArticlesParams) (*SearchArticlesResult, error) {
			ctx, cancel := svcCtx.ToolTimeout(ctx)
			defer cancel()
			if params.PageSize <= 0 || params.PageSize > 20 {
				params.PageSize = 10
			}
			resp, err := svcCtx.Article.SearchArticles(ctx, &article.SearchRequest{
				Keyword:  params.Keyword,
				PageSize: int64(params.PageSize),
			})
			if err != nil {
				return nil, err
			}
			items := make([]SearchArticleItem, 0, len(resp.Items))
			for _, a := range resp.Items {
				items = append(items, SearchArticleItem{
					Title: a.Title, Description: a.Description,
					AuthorName: a.AuthorName, LikeNum: a.LikeNum,
					CommentNum: a.CommentNum, PublishTime: a.PublishTime,
				})
			}
			return &SearchArticlesResult{Articles: items, Total: int64(len(items))}, nil
		})
}
