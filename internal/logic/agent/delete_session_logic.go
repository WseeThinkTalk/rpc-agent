package agentlogic

import (
	"context"

	"rpc-agent/agent"
	"rpc-agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteSessionLogic {
	return &DeleteSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteSessionLogic) DeleteSession(in *agent.DeleteSessionRequest) (resp *agent.DeleteSessionResponse, err error) {
	resp = new(agent.DeleteSessionResponse)
	resp.Data = new(agent.DeleteSessionData)

	if err = l.svcCtx.SessMgr.DeleteSession(l.ctx, in.UserId, in.SessionId); err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data.Success = false
		return resp, err
	}

	resp.Data.Success = true
	return resp, nil
}
