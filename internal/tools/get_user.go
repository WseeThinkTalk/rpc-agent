package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/internal/svc"
	"rpc-agent/client/user/user"
)

type GetUserInfoParams struct {
	Username string `json:"username" description:"用户昵称/用户名/账号名，进行模糊匹配查找"`
}

type UserInfoResult struct {
	Username string `json:"username"`
	Bio      string `json:"bio"`
}

func NewGetUserInfoTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("get_user_info",
		"根据用户名/昵称模糊搜索获取指定用户的公开信息，包含用户名及简介。",
		func(ctx context.Context, params *GetUserInfoParams) (*UserInfoResult, error) {
			ctx, cancel := svcCtx.ToolTimeout(ctx)
			defer cancel()
			resp, err := svcCtx.User.AdminUserList(ctx, &user.AdminUserListRequest{
				Keyword:  params.Username,
				PageSize: 1,
			})
			if err != nil {
				return nil, err
			}
			if len(resp.Items) == 0 {
				return nil, fmt.Errorf("未找到昵称或账号匹配为 '%s' 的用户", params.Username)
			}
			u := resp.Items[0]
			return &UserInfoResult{
				Username: u.Username,
				Bio:      u.Bio,
			}, nil
		})
}
