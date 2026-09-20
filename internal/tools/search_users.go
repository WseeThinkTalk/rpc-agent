package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/internal/svc"
	"rpc-agent/client/user/user"
)

type SearchUsersParams struct {
	Keyword  string `json:"keyword" description:"搜索关键词，匹配用户昵称、个人简介或账号ID（DisplayID）"`
	PageSize int32  `json:"page_size" description:"返回数量，1-20"`
}

type SearchUserItem struct {
	Username  string `json:"username"`
	Bio       string `json:"bio"`
}

type SearchUsersResult struct {
	Users []SearchUserItem `json:"users"`
	Total int64            `json:"total"`
}

func NewSearchUsersTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("search_users",
		"在 ThinkTalk 平台搜索注册用户。支持按昵称、简介进行模糊查询匹配，返回匹配的用户列表信息。",
		func(ctx context.Context, params *SearchUsersParams) (*SearchUsersResult, error) {
			ctx, cancel := svcCtx.ToolTimeout(ctx)
			defer cancel()
			if params.PageSize <= 0 || params.PageSize > 20 {
				params.PageSize = 10
			}
			resp, err := svcCtx.User.AdminUserList(ctx, &user.AdminUserListRequest{
				Keyword:  params.Keyword,
				PageSize: int64(params.PageSize),
			})
			if err != nil {
				return nil, err
			}
			items := make([]SearchUserItem, 0, len(resp.Items))
			for _, u := range resp.Items {
				items = append(items, SearchUserItem{
					Username:  u.Username,
					Bio:       u.Bio,
				})
			}
			return &SearchUsersResult{Users: items, Total: int64(len(items))}, nil
		})
}
