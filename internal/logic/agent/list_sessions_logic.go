package agentlogic

import (
	"context"

	"rpc-agent/agent"
	"rpc-agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSessionsLogic {
	return &ListSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSessionsLogic) ListSessions(in *agent.ListSessionsRequest) (*agent.ListSessionsResponse, error) {
	summaries, err := l.svcCtx.SessMgr.ListSessions(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	pbSessions := make([]*agent.SessionSummary, len(summaries))
	for i, sum := range summaries {
		pbSessions[i] = &agent.SessionSummary{
			SessionId:    sum.SessionID,
			Title:        sum.Title,
			MessageCount: sum.MessageCount,
			CreatedAt:    sum.CreatedAt,
			UpdatedAt:    sum.UpdatedAt,
		}
	}
	return &agent.ListSessionsResponse{
		Code: 200,
		Msg:  "success",
		Data: &agent.ListSessionsData{
			Sessions: pbSessions,
		},
	}, nil
}
