package agentlogic

import (
	"context"

	"rpc-agent/agent"
	"rpc-agent/internal/model"
	"rpc-agent/internal/svc"
	"rpc-agent/pkg/code"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHistoryLogic {
	return &GetHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetHistoryLogic) GetHistory(in *agent.GetHistoryRequest) (resp *agent.GetHistoryResponse, err error) {
	resp = new(agent.GetHistoryResponse)
	resp.Data = new(agent.GetHistoryData)
	resp.Data.Messages = make([]*agent.HistoryMessage, 0)

	sess, err := l.svcCtx.SessMgr.Load(l.ctx, in.UserId, in.SessionId)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}

	// 转换会话历史消息为响应 DTO
	for _, v := range sess.Messages {
		resp.Data.Messages = append(resp.Data.Messages, &agent.HistoryMessage{Role: v.Role, Content: v.Content})
	}
	title := sess.Title
	if title == "" {
		// 从首条用户输入提取会话标题
		for _, v := range sess.Messages {
			if v.Role == "user" {
				title = model.TruncateRunes(v.Content, 30)
				break
			}
		}
	}

	resp.Data.SessionId = in.SessionId
	resp.Data.Title = title
	resp.Data.CreatedAt = sess.CreatedAt
	resp.Data.UpdatedAt = sess.UpdatedAt

	return resp, nil
}
