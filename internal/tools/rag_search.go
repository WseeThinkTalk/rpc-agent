package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"rpc-agent/client/article/article"
	"rpc-agent/internal/svc"
	"rpc-agent/pkg/embedding"
	"rpc-agent/pkg/hybrid"
)

// RAGSearchParams 语义搜索参数
type RAGSearchParams struct {
	Query string `json:"query" description:"用户的自然语言问题或搜索描述"`
	TopK  int    `json:"top_k" description:"返回最相关的文章数量，1-10，默认 5"`
}

// RAGSearchItem 单篇相关文章
type RAGSearchItem struct {
	ArticleID    int64   `json:"article_id"`
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

// NewRAGSearchTool 创建语义+关键词 RRF 混合搜索工具（RAG 检索）
// 当 DB 或 EmbedClient 未配置时，返回 nil, nil（由 registry 跳过注册）
func NewRAGSearchTool(svcCtx *svc.ServiceContext) (tool.InvokableTool, error) {
	if svcCtx.DB == nil || svcCtx.EmbedClient == nil {
		return nil, nil
	}

	return utils.InferTool(
		"semantic_search_articles",
		"基于语义相似度与倒排关键字 RRF 混合检索平台文章。当用户提出概念性问题、想了解某个主题、或关键词搜索效果不佳时使用此工具。"+
			"该工具并行检索 ES 倒排索引与 pgvector 语义向量库，通过 RRF (Reciprocal Rank Fusion) 融合出最相关的文章片段。",
		func(ctx context.Context, params *RAGSearchParams) (*RAGSearchResult, error) {
			if svcCtx.DB == nil || svcCtx.EmbedClient == nil {
				return &RAGSearchResult{
					Message: "RAG 功能未配置，请使用普通搜索代替",
				}, nil
			}

			topK := params.TopK
			if topK <= 0 || topK > 10 {
				topK = 5
			}

			var (
				wg            sync.WaitGroup
				vecArticleIDs []int64
				vecItemMap    = make(map[int64]RAGSearchItem)
				esArticleIDs  []int64
				esItemMap     = make(map[int64]RAGSearchItem)
				vecErr, esErr error
			)

			// 1. 并发支路 A：pgvector 语义向量检索 (HNSW 索引 + 会话级参数调优 + 余弦相似度)
			wg.Add(1)
			go func() {
				defer wg.Done()
				queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				vec, err := svcCtx.EmbedClient.Embed(queryCtx, params.Query)
				if err != nil || len(vec) == 0 {
					vecErr = err
					return
				}

				// 会话级动态设置 HNSW 搜索宽度，平衡召回率与毫秒级延迟
				_, _ = svcCtx.DB.ExecContext(ctx, "SET LOCAL hnsw.ef_search = 40")

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
				`, vecStr, topK*2)
				if err != nil {
					vecErr = err
					return
				}
				defer rows.Close()

				for rows.Next() {
					var item RAGSearchItem
					if err := rows.Scan(&item.ArticleID, &item.Title, &item.ContentChunk, &item.Similarity); err == nil {
						if item.Similarity >= 0.3 {
							vecArticleIDs = append(vecArticleIDs, item.ArticleID)
							vecItemMap[item.ArticleID] = item
						}
					}
				}
			}()

			// 2. 并发支路 B：ES 倒排索引关键词检索 (通过 Article RPC 跨服务调用)
			if svcCtx.Article != nil {
				wg.Add(1)
				go func() {
					defer wg.Done()
					searchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()

					res, err := svcCtx.Article.SearchArticles(searchCtx, &article.SearchRequest{
						Keyword:  params.Query,
						PageSize: int64(topK * 2),
					})
					if err != nil || res == nil || len(res.Items) == 0 {
						esErr = err
						return
					}

					for _, it := range res.Items {
						esArticleIDs = append(esArticleIDs, it.ArticleId)
						esItemMap[it.ArticleId] = RAGSearchItem{
							ArticleID:    it.ArticleId,
							Title:        it.Title,
							ContentChunk: it.Description,
							Similarity:   0.8,
						}
					}
				}()
			}

			wg.Wait()

			// 3. 执行工业级 RRF (Reciprocal Rank Fusion) 倒数排名融合
			var rankLists [][]int64
			if len(vecArticleIDs) > 0 {
				rankLists = append(rankLists, vecArticleIDs)
			}
			if len(esArticleIDs) > 0 {
				rankLists = append(rankLists, esArticleIDs)
			}

			if len(rankLists) == 0 {
				return &RAGSearchResult{
					Articles: []RAGSearchItem{},
					Total:    0,
					Message:  fmt.Sprintf("未检索到相关内容 (vecErr: %v, esErr: %v)", vecErr, esErr),
				}, nil
			}

			fused := hybrid.ReciprocalRankFusion(rankLists, hybrid.DefaultRRFK)
			if len(fused) > topK {
				fused = fused[:topK]
			}

			// 4. 组装融合输出
			finalItems := make([]RAGSearchItem, 0, len(fused))
			for _, f := range fused {
				if item, ok := vecItemMap[f.ID]; ok {
					item.Similarity = f.RRFScore
					finalItems = append(finalItems, item)
				} else if item, ok := esItemMap[f.ID]; ok {
					item.Similarity = f.RRFScore
					finalItems = append(finalItems, item)
				}
			}

			return &RAGSearchResult{
				Articles: finalItems,
				Total:    len(finalItems),
			}, nil
		},
	)
}
