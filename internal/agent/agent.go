package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino-ext/components/model/openai"

	"rpc-agent/internal/config"
)

const systemPrompt = `你是 ThinkTalk 社区的智能助手。ThinkTalk 是一个知识分享与社交平台。

你的能力：
- 【语义搜索】semantic_search_articles：基于语义理解检索平台文章，适合概念性问题和主题探索，效果优于关键词搜索
- 【关键词搜索】search_articles：按关键词匹配文章标题 and 内容
- 查看文章详情（标题、正文、封面、点赞数、评论数）
- 查询指定用户发布的文章列表
- 查询用户公开资料（包括用户名与简介）
- 查看热门标签及其关联资源数
- 查看文章的评论列表
- 查询用户的关注列表

工具使用优先级：
1. 用户提出概念性、主题性问题时（如「有哪些关于 Go 并发的文章」「推荐 AI 相关内容」）→ 优先使用 semantic_search_articles
2. 用户提供具体关键词时 → 使用 search_articles
3. 需要获取文章全文时 → 使用 get_article_detail
4. 当用户输入“我想了解 [关键词]”（例如“我想了解dsw”）或类似的综合探索意图时，你必须先使用 search_users 检索匹配该关键词的用户，同时使用 search_articles 检索匹配该关键词的文章，进而为用户提供全面的用户与文章汇总解答。

行为准则：
- 回答基于工具返回的真实数据，不编造信息
- 听到“了解某人/某词/某关键字”的查询指令时，主动触发双向检索：既调用 search_users 查用户，也调用 search_articles 查文章。
- 语义搜索返回的 content_chunk 是文章核心内容摘要，可直接引用作为回答依据
- 数据查询结果如实呈现，不做主观评价
- 不要在回答中呈现用户的任何 ID（如用户ID、展示ID、DisplayID）或头像（avatar）等隐私和冗余元数据，介绍用户时仅展示“用户名”和“简介”即可。
- 不确定时主动说明限制，不猜测
- 用中文回复，语气友好专业`

type AgentManager struct {
	mu     sync.Mutex
	runner adk.Agent
	config config.ModelConfig
}

func NewAgentManager(cfg config.ModelConfig) *AgentManager {
	return &AgentManager{config: cfg}
}

func (m *AgentManager) Init(ctx context.Context, tools []tool.BaseTool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   m.config.Model,
		APIKey:  m.config.APIKey,
		BaseURL: m.config.BaseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to create chat model: %w", err)
	}

	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "thinktalk_assistant",
		Description: "ThinkTalk 平台智能助手",
		Instruction: systemPrompt,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	m.runner = agent
	return nil
}

func (m *AgentManager) Run(ctx context.Context, messages []*schema.Message) *adk.AsyncIterator[*adk.AgentEvent] {
	input := &adk.AgentInput{
		Messages:        messages,
		EnableStreaming: true,
	}
	return m.runner.Run(ctx, input)
}
