package hybrid

import "sort"

// DefaultRRFK 工业界标准平滑参数 k（通常取 60，避免头部差距过大）
const DefaultRRFK = 60

// FusedResult 融合排序结果项
type FusedResult struct {
	ID       int64
	RRFScore float64
}

// ReciprocalRankFusion 实现倒数排名融合算法（Reciprocal Rank Fusion）
// 公式：RRF_Score(d) = sum_{s in sources} 1 / (k + rank_s(d))
// 输入为各个检索路（如 ES 关键字排位、pgvector 语义排位）的有序 ID 切片
// 输出为综合加权打分后的融合降序切片
func ReciprocalRankFusion(rankings [][]int64, k int) []FusedResult {
	if k <= 0 {
		k = DefaultRRFK
	}

	scores := make(map[int64]float64)
	for _, rankList := range rankings {
		for rank, id := range rankList {
			// rank 为 0-indexed，转换为公式中 1-indexed: (rank + 1)
			scores[id] += 1.0 / float64(k+rank+1)
		}
	}

	results := make([]FusedResult, 0, len(scores))
	for id, score := range scores {
		results = append(results, FusedResult{
			ID:       id,
			RRFScore: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].RRFScore == results[j].RRFScore {
			return results[i].ID > results[j].ID
		}
		return results[i].RRFScore > results[j].RRFScore
	})

	return results
}
