package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/agent/internal/svc"
	"rpc-agent/client/article/article"
	"rpc-agent/client/user/user"
)

type GetUserArticlesParams struct {
	Username string `json:"username" description:"作者用户名或昵称，模糊匹配定位作者"`
	PageSize int32  `json:"page_size" description:"返回数量，1-20"`
}

type UserArticleItem struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	LikeCount    int64  `json:"like_count"`
	CommentCount int64  `json:"comment_count"`
	PublishTime  int64  `json:"publish_time"`
}

type GetUserArticlesResult struct {
	Articles []UserArticleItem `json:"articles"`
	Total    int64             `json:"total"`
}

func NewGetUserArticlesTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("get_user_articles",
		"获取指定用户发布的文章列表。可按作者的用户名/昵称模糊搜索查询。",
		func(ctx context.Context, params *GetUserArticlesParams) (*GetUserArticlesResult, error) {
			ctx, cancel := svcCtx.ToolTimeout(ctx)
			defer cancel()
			if params.PageSize <= 0 || params.PageSize > 20 {
				params.PageSize = 10
			}
			userResp, err := svcCtx.User.AdminUserList(ctx, &user.AdminUserListRequest{
				Keyword:  params.Username,
				PageSize: 1,
			})
			if err != nil {
				return nil, err
			}
			if len(userResp.Items) == 0 {
				return nil, fmt.Errorf("未找到昵称或账号匹配为 '%s' 的博主", params.Username)
			}
			authorID := userResp.Items[0].UserId

			resp, err := svcCtx.Article.Articles(ctx, &article.ArticlesRequest{
				UserId:   authorID,
				PageSize: int64(params.PageSize),
			})
			if err != nil {
				return nil, err
			}
			items := make([]UserArticleItem, 0, len(resp.Articles))
			for _, a := range resp.Articles {
				items = append(items, UserArticleItem{
					Title: a.Title, Description: a.Description,
					LikeCount: a.LikeCount, CommentCount: a.CommentCount,
					PublishTime: a.PublishTime,
				})
			}
			return &GetUserArticlesResult{Articles: items, Total: int64(len(items))}, nil
		})
}
