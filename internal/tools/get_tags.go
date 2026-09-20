package tools

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/internal/svc"
	"rpc-agent/client/tag/tag"
)

type GetHotTagsParams struct {
	Limit int32 `json:"limit" description:"返回数量，1-20"`
}

type TagItemResult struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	ResourceCount int64  `json:"resource_count"`
}

type GetHotTagsResult struct {
	Tags []TagItemResult `json:"tags"`
}

func NewGetHotTagsTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	return utils.InferTool("get_hot_tags",
		"获取 ThinkTalk 平台上的热门标签列表，每个标签含关联资源数量。",
		func(ctx context.Context, params *GetHotTagsParams) (*GetHotTagsResult, error) {
			ctx, cancel := svcCtx.ToolTimeout(ctx)
			defer cancel()
			if params.Limit <= 0 || params.Limit > 20 {
				params.Limit = 10
			}
			resp, err := svcCtx.Tag.HotTags(ctx, &tag.HotTagsRequest{Limit: params.Limit})
			if err != nil {
				return nil, err
			}
			items := make([]TagItemResult, 0, len(resp.Items))
			for _, t := range resp.Items {
				items = append(items, TagItemResult{
					ID: t.TagId, Name: t.TagName, Desc: t.TagDesc, ResourceCount: t.ResourceCount,
				})
			}
			return &GetHotTagsResult{Tags: items}, nil
		})
}
