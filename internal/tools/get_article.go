package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/internal/svc"
	"rpc-agent/client/article/article"
	"rpc-agent/client/user/user"
)

type GetArticleParams struct {
	Keyword string `json:"keyword" description:"文章的标题或关键内容，进行模糊匹配查找"`
}

type ArticleDetailResult struct {
	Title        string `json:"title"`
	Content      string `json:"content"`
	Description  string `json:"description"`
	AuthorName   string `json:"author_name"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	PublishTime  int64  `json:"publish_time"`
}

func NewGetArticleDetailTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("get_article_detail",
		"模糊匹配标题或内容获取指定文章的完整详情，包括标题、正文、作者姓名、点赞数、评论数等。",
		func(ctx context.Context, params *GetArticleParams) (*ArticleDetailResult, error) {
			ctx, cancel := svcCtx.ToolTimeout(ctx)
			defer cancel()
			searchResp, err := svcCtx.Article.SearchArticles(ctx, &article.SearchRequest{
				Keyword:  params.Keyword,
				PageSize: 1,
			})
			if err != nil {
				return nil, err
			}
			if len(searchResp.Items) == 0 {
				return nil, fmt.Errorf("未找到标题或内容匹配为 '%s' 的文章", params.Keyword)
			}
			articleID := searchResp.Items[0].ArticleId

			resp, err := svcCtx.Article.ArticleDetail(ctx, &article.ArticleDetailRequest{
				ArticleId: articleID,
			})
			if err != nil {
				return nil, err
			}
			a := resp.Article
			if a == nil {
				return nil, nil
			}

			// 获取作者昵称
			authorName := "未知用户"
			if userResp, err := svcCtx.User.FindById(ctx, &user.FindByIdRequest{UserId: a.AuthorId}); err == nil && userResp != nil {
				authorName = userResp.Username
			}

			return &ArticleDetailResult{
				Title: a.Title, Content: a.Content,
				Description: a.Description, AuthorName: authorName,
				LikeCount: a.LikeCount, CommentCount: a.CommentCount,
				PublishTime: a.PublishTime,
			}, nil
		})
}
