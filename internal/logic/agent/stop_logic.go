package agentlogic

import (
	"context"

	"rpc-agent/agent"
	"rpc-agent/internal/model"
	"rpc-agent/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type StopLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStopLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StopLogic {
	return &StopLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *StopLogic) Stop(in *agent.StopRequest) (*agent.StopResponse, error) {
	sessionKey := model.SessionKey(in.UserId, in.SessionId)
	l.svcCtx.StopSession(sessionKey)
	return &agent.StopResponse{
		Code: 200,
		Msg:  "success",
	}, nil
}
