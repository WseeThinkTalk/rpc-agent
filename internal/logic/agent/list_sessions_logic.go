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

func (l *ListSessionsLogic) ListSessions(in *agent.ListSessionsRequest) (resp *agent.ListSessionsResponse, err error) {
	resp = new(agent.ListSessionsResponse)
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = new(agent.ListSessionsData)
	resp.Data.Sessions = make([]*agent.SessionSummary, 0)

	summaries, err := l.svcCtx.SessMgr.ListSessions(l.ctx, in.UserId)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		return resp, nil
	}

	for _, sum := range summaries {
		resp.Data.Sessions = append(resp.Data.Sessions, &agent.SessionSummary{
			SessionId:    sum.SessionID,
			Title:        sum.Title,
			MessageCount: sum.MessageCount,
			CreatedAt:    sum.CreatedAt,
			UpdatedAt:    sum.UpdatedAt,
		})
	}

	return resp, nil
}
