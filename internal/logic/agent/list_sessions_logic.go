package agentlogic

import (
	"context"

	"rpc-agent/agent"
	"rpc-agent/internal/svc"
	"rpc-agent/pkg/code"

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
	resp.Data = new(agent.ListSessionsData)
	resp.Data.Sessions = make([]*agent.SessionSummary, 0)

	summaries, err := l.svcCtx.SessMgr.ListSessions(l.ctx, in.UserId)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}

	// 组装历史会话概览列表
	for _, v := range summaries {
		resp.Data.Sessions = append(resp.Data.Sessions, &agent.SessionSummary{
			SessionId:    v.SessionID,
			Title:        v.Title,
			MessageCount: v.MessageCount,
			CreatedAt:    v.CreatedAt,
			UpdatedAt:    v.UpdatedAt,
		})
	}

	return resp, nil
}
