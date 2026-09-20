package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/internal/svc"
	"rpc-agent/client/article/article"
	"rpc-agent/client/reply/reply"
	"rpc-agent/client/user/user"
)

type GetArticleCommentsParams struct {
	Keyword  string `json:"keyword" description:"文章的标题或关键内容，进行模糊匹配查找"`
	PageSize int32  `json:"page_size" description:"返回数量，1-20"`
}

type CommentItem struct {
	Content       string `json:"content"`
	ReplyUsername string `json:"reply_username"`
	LikeNum       int64  `json:"like_num"`
	CreateTime    int64  `json:"create_time"`
}

type GetArticleCommentsResult struct {
	Comments []CommentItem `json:"comments"`
	Total    int64         `json:"total"`
}

func NewGetArticleCommentsTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("get_article_comments",
		"获取指定文章的评论列表。可输入文章标题或关键内容进行模糊搜索匹配。",
		func(ctx context.Context, params *GetArticleCommentsParams) (*GetArticleCommentsResult, error) {
			ctx, cancel := svcCtx.ToolTimeout(ctx)
			defer cancel()
			if params.PageSize <= 0 || params.PageSize > 20 {
				params.PageSize = 10
			}
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

			resp, err := svcCtx.Reply.ReplyList(ctx, &reply.ReplyListRequest{
				BizId: "article", TargetId: articleID, PageSize: int64(params.PageSize),
			})
			if err != nil {
				return nil, err
			}
			items := make([]CommentItem, 0, len(resp.Items))
			for _, r := range resp.Items {
				commenterName := "未知用户"
				if uResp, err := svcCtx.User.FindById(ctx, &user.FindByIdRequest{UserId: r.ReplyUserId}); err == nil && uResp != nil {
					commenterName = uResp.Username
				}
				items = append(items, CommentItem{
					Content:       r.Content,
					ReplyUsername: commenterName,
					LikeNum:       r.LikeNum,
					CreateTime:    r.CreateTime,
				})
			}
			return &GetArticleCommentsResult{Comments: items, Total: int64(len(items))}, nil
		})
}
