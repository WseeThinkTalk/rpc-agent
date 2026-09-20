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

func (l *GetHistoryLogic) GetHistory(in *agent.GetHistoryRequest) (*agent.GetHistoryResponse, error) {
	sess, err := l.svcCtx.SessMgr.Load(l.ctx, in.UserId, in.SessionId)
	if err != nil {
		return nil, err
	}
	msgs := make([]*agent.HistoryMessage, len(sess.Messages))
	for i, m := range sess.Messages {
		msgs[i] = &agent.HistoryMessage{Role: m.Role, Content: m.Content}
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
	return &agent.GetHistoryResponse{
		Code: 200,
		Msg:  "success",
		Data: &agent.GetHistoryData{
			SessionId: in.SessionId,
			Title:     title,
			Messages:  msgs,
			CreatedAt: sess.CreatedAt,
			UpdatedAt: sess.UpdatedAt,
		},
	}, nil
}
