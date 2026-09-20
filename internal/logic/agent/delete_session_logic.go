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

func (l *DeleteSessionLogic) DeleteSession(in *agent.DeleteSessionRequest) (*agent.DeleteSessionResponse, error) {
	if err := l.svcCtx.SessMgr.DeleteSession(l.ctx, in.UserId, in.SessionId); err != nil {
		return &agent.DeleteSessionResponse{
			Code: 500,
			Msg:  err.Error(),
			Data: &agent.DeleteSessionData{Success: false},
		}, err
	}
	return &agent.DeleteSessionResponse{
		Code: 200,
		Msg:  "success",
		Data: &agent.DeleteSessionData{Success: true},
	}, nil
}
