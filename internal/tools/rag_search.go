package tools

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/internal/svc"
	"rpc-agent/pkg/embedding"
)

// RAGSearchParams 语义搜索参数
type RAGSearchParams struct {
	Query   string `json:"query" description:"用户的自然语言问题或搜索描述"`
	TopK    int    `json:"top_k" description:"返回最相关的文章数量，1-10，默认 5"`
}

// RAGSearchItem 单篇相关文章
type RAGSearchItem struct {
	Title        string  `json:"title"`
	ContentChunk string  `json:"content_chunk"`
	Similarity   float64 `json:"similarity"`
}

// RAGSearchResult 语义搜索结果
type RAGSearchResult struct {
	Articles []RAGSearchItem `json:"articles"`
	Total    int             `json:"total"`
	Message  string          `json:"message,omitempty"`
}

// NewRAGSearchTool 创建语义搜索工具（RAG 检索）
// 当 DB 或 EmbedClient 未配置时，返回 nil, nil（由 registry 跳过注册）
func NewRAGSearchTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	if svcCtx.DB == nil || svcCtx.EmbedClient == nil {
		return nil, nil
	}

	return utils.InferTool(
		"semantic_search_articles",
		"基于语义相似度搜索平台文章。当用户提出概念性问题、想了解某个主题、或关键词搜索效果不佳时使用此工具。"+
			"该工具会将用户问题转为向量，在文章库中找出内容最相关的文章，适合回答「有哪些关于X的文章」「推荐一些Y主题的内容」等问题。",
		func(ctx context.Context, params *RAGSearchParams) (*RAGSearchResult, error) {
			if svcCtx.DB == nil || svcCtx.EmbedClient == nil {
				return &RAGSearchResult{
					Message: "RAG 功能未配置，请使用 search_articles 工具代替",
				}, nil
			}

			topK := params.TopK
			if topK <= 0 || topK > 10 {
				topK = 5
			}

			// 1. 将查询转换为向量
			queryCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			vec, err := svcCtx.EmbedClient.Embed(queryCtx, params.Query)
			if err != nil {
				return nil, fmt.Errorf("生成查询向量失败: %w", err)
			}
			if len(vec) == 0 {
				return nil, fmt.Errorf("向量化结果为空")
			}

			// 2. 在 pgvector 中做余弦相似度检索
			vecStr := embedding.VectorToSQL(vec)
			rows, err := svcCtx.DB.QueryContext(ctx, `
				SELECT
					article_id,
					title,
					content_chunk,
					1 - (embedding <=> $1::vector) AS similarity
				FROM article_embeddings
				ORDER BY embedding <=> $1::vector
				LIMIT $2
			`, vecStr, topK)
			if err != nil {
				return nil, fmt.Errorf("向量检索失败: %w", err)
			}
			defer rows.Close()

			var items []RAGSearchItem
			for rows.Next() {
				var articleID int64
				var item RAGSearchItem
				if err := rows.Scan(&articleID, &item.Title, &item.ContentChunk, &item.Similarity); err != nil {
					continue
				}
				// 只返回相似度 > 0.4 的结果
				if item.Similarity < 0.4 {
					continue
				}
				items = append(items, item)
			}

			if len(items) == 0 {
				return &RAGSearchResult{
					Articles: []RAGSearchItem{},
					Total:    0,
					Message:  "未找到与该问题高度相关的文章，请尝试 search_articles 工具或换一种描述",
				}, nil
			}

			return &RAGSearchResult{
				Articles: items,
				Total:    len(items),
			}, nil
		},
	)
}
