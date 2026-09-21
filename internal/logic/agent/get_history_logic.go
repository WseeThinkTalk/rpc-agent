package agentlogic

import (
	"context"

	"rpc-agent/agent"
	"rpc-agent/internal/model"
	"rpc-agent/internal/svc"

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
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = new(agent.GetHistoryData)
	resp.Data.Messages = make([]*agent.HistoryMessage, 0)

	sess, err := l.svcCtx.SessMgr.Load(l.ctx, in.UserId, in.SessionId)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, nil
	}

	for _, m := range sess.Messages {
		resp.Data.Messages = append(resp.Data.Messages, &agent.HistoryMessage{Role: m.Role, Content: m.Content})
	}
	title := sess.Title
	if title == "" {
		for _, m := range sess.Messages {
			if m.Role == "user" {
				title = model.TruncateRunes(m.Content, 30)
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
