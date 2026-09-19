package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/agent/internal/svc"
	"rpc-agent/client/follow/follow"
	"rpc-agent/client/user/user"
)

type GetUserFollowersParams struct {
	Username string `json:"username" description:"需要查询的用户的用户名或昵称"`
	PageSize int32  `json:"page_size" description:"返回数量，1-20"`
}

type FollowItemResult struct {
	FollowedUsername string `json:"followed_username"`
	FansCount        int64  `json:"fans_count"`
	CreateTime       int64  `json:"create_time"`
}

type GetUserFollowersResult struct {
	FollowList []FollowItemResult `json:"follow_list"`
	Total      int64              `json:"total"`
}

func NewGetUserFollowersTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("get_user_followers",
		"模糊匹配用户名/昵称，获取指定用户的关注列表，了解用户关注了谁及被关注者的粉丝数。",
		func(ctx context.Context, params *GetUserFollowersParams) (*GetUserFollowersResult, error) {
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
				return nil, fmt.Errorf("未找到昵称或账号匹配为 '%s' 的用户", params.Username)
			}
			userID := userResp.Items[0].UserId

			resp, err := svcCtx.Follow.FollowList(ctx, &follow.FollowListRequest{
				UserId: userID, PageSize: int64(params.PageSize),
			})
			if err != nil {
				return nil, err
			}
			items := make([]FollowItemResult, 0, len(resp.Items))
			for _, f := range resp.Items {
				followedName := "未知用户"
				if uResp, err := svcCtx.User.FindById(ctx, &user.FindByIdRequest{UserId: f.FollowedUserId}); err == nil && uResp != nil {
					followedName = uResp.Username
				}
				items = append(items, FollowItemResult{
					FollowedUsername: followedName, FansCount: f.FansCount, CreateTime: f.CreateTime,
				})
			}
			return &GetUserFollowersResult{FollowList: items, Total: int64(len(items))}, nil
		})
}
