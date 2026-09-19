// Package embedding 封装 Doubao Embedding API 调用（OpenAI 兼容接口）
package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL   = "https://ark.cn-beijing.volces.com/api/v3"
	DefaultDimension = 1024
)

// Client 是 Doubao Embedding 客户端
type Client struct {
	apiKey     string
	model      string
	baseURL    string
	dimensions int
	http       *http.Client
}

// Config 配置
type Config struct {
	APIKey     string
	Model      string
	BaseURL    string // 可选，默认使用 DefaultBaseURL
	Dimensions int    // 可选，指定向量维度（如 1024），0 表示使用模型默认值
}

type embeddingRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient 创建新的 Embedding 客户端
func NewClient(cfg Config) *Client {
	base := cfg.BaseURL
	if base == "" {
		base = DefaultBaseURL
	}
	dim := cfg.Dimensions
	if dim == 0 {
		dim = DefaultDimension
	}
	return &Client{
		apiKey:     cfg.APIKey,
		model:      cfg.Model,
		baseURL:    base,
		dimensions: dim,
		http:       &http.Client{Timeout: 30 * time.Second},
	}
}

// Embed 对单条文本生成 embedding 向量
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	vecs, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("embedding: empty response")
	}
	return vecs[0], nil
}

// EmbedBatch 批量生成 embedding 向量（最多 32 条）
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	body, err := json.Marshal(embeddingRequest{
		Model:      c.model,
		Input:      texts,
	})
	if err != nil {
		return nil, fmt.Errorf("embedding: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("embedding: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding: http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("embedding: read response: %w", err)
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, fmt.Errorf("embedding: unmarshal response: %w, body: %s", err, string(respBody))
	}
	if embResp.Error != nil {
		return nil, fmt.Errorf("embedding API error [%s]: %s", embResp.Error.Code, embResp.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding: API returned %d: %s", resp.StatusCode, string(respBody))
	}

	// 按 index 排序结果
	result := make([][]float32, len(texts))
	for _, d := range embResp.Data {
		if d.Index < len(result) {
			result[d.Index] = d.Embedding
		}
	}
	return result, nil
}

// VectorToSQL 将 []float32 转换为 pgvector 可识别的字符串格式 "[f1,f2,...]"
func VectorToSQL(v []float32) string {
	b := make([]byte, 0, len(v)*12)
	b = append(b, '[')
	for i, f := range v {
		if i > 0 {
			b = append(b, ',')
		}
		b = fmt.Appendf(b, "%g", f)
	}
	b = append(b, ']')
	return string(b)
}
