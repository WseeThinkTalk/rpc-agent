package tools

import (
	"github.com/cloudwego/eino/components/tool"

	"rpc-agent/agent/internal/svc"
)

func All(svcCtx *svc.ServiceContext) []tool.BaseTool {
	invokable := []tool.InvokableTool{
		must(NewSearchArticlesTool(svcCtx)),
		must(NewSearchUsersTool(svcCtx)),
		must(NewGetArticleDetailTool(svcCtx)),
		must(NewGetUserArticlesTool(svcCtx)),
		must(NewGetUserInfoTool(svcCtx)),
		must(NewGetHotTagsTool(svcCtx)),
		must(NewGetArticleCommentsTool(svcCtx)),
		must(NewGetUserFollowersTool(svcCtx)),
	}

	// RAG 语义搜索工具（可选：DB 和 EmbedClient 均配置时才注册）
	ragTool, err := NewRAGSearchTool(svcCtx)
	if err != nil {
		panic("failed to init RAG search tool: " + err.Error())
	}
	if ragTool != nil {
		invokable = append(invokable, ragTool)
	}

	result := make([]tool.BaseTool, len(invokable))
	for i, t := range invokable {
		result[i] = t
	}
	return result
}

func must(t tool.InvokableTool, err error) tool.InvokableTool {
	if err != nil {
		panic(err)
	}
	return t
}
