package agentlogic

import (
	"context"
	"io"
	"time"

	"rpc-agent/agent"
	"rpc-agent/internal/model"
	"rpc-agent/internal/svc"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ChatLogic) Chat(req *agent.ChatRequest, stream agent.Agent_ChatServer) error {
	ctx := stream.Context()
	l.Infof("[agent-rpc] Chat called: userId=%d sessionId=%s msg=%s", req.UserId, req.SessionId, req.Message)

	sessionKey := model.SessionKey(req.UserId, req.SessionId)

	sess, _ := l.svcCtx.SessMgr.Load(ctx, req.UserId, req.SessionId)
	sess.Messages = append(sess.Messages, model.Message{Role: "user", Content: req.Message})

	// 裁剪上下文
	const maxContextMessages = 20
	ctxMessages := sess.Messages
	if len(ctxMessages) > maxContextMessages {
		ctxMessages = ctxMessages[len(ctxMessages)-maxContextMessages:]
	}

	// 可取消的 context：用户点 Stop 或连接断开时立即中断
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stopCh, cleanup := l.svcCtx.RegisterStopChan(sessionKey)
	defer cleanup()

	go func() {
		select {
		case <-stopCh:
			cancel()
		case <-runCtx.Done():
		}
	}()

	iter := l.svcCtx.AgentMgr.Run(runCtx, model.MessagesToSchema(ctxMessages))
	logx.WithContext(runCtx).Debugf("[agent-rpc] Agent.Run returned iterator, starting event loop")

	eventCh := make(chan *adk.AgentEvent, 8)
	go func() {
		defer close(eventCh)
		for {
			event, ok := iter.Next()
			if !ok {
				return
			}
			select {
			case eventCh <- event:
			case <-runCtx.Done():
				return
			}
		}
	}()

	heartbeat := time.NewTicker(3 * time.Second)
	defer heartbeat.Stop()

	var fullReply string
	stopped := false

	for {
		select {
		case <-heartbeat.C:
			_ = stream.Send(&agent.ChatEvent{Type: "thought", Content: "正在思考..."})

		case <-runCtx.Done():
			_ = stream.Send(&agent.ChatEvent{Type: "done", FinishReason: "stopped"})
			return nil

		case event, ok := <-eventCh:
			if !ok {
				goto saveAndDone
			}
			heartbeat.Reset(3 * time.Second)

			logx.WithContext(runCtx).Debugf("[agent-rpc] got event: err=%v output=%v action=%v",
				event.Err != nil,
				event.Output != nil && event.Output.MessageOutput != nil,
				event.Action != nil)

			l.processEvent(runCtx, stopCh, event, stream, &fullReply, &stopped)
			if stopped {
				_ = stream.Send(&agent.ChatEvent{Type: "done", FinishReason: "stopped"})
				return nil
			}
		}
	}

saveAndDone:
	sess.Messages = append(sess.Messages, model.Message{Role: "assistant", Content: fullReply})
	l.svcCtx.SessMgr.Save(ctx, req.UserId, req.SessionId, sess)

	stream.Send(&agent.ChatEvent{Type: "done", FinishReason: "stop"})
	return nil
}

func (l *ChatLogic) processEvent(ctx context.Context, stopCh chan struct{}, event *adk.AgentEvent, stream agent.Agent_ChatServer, fullReply *string, stopped *bool) {
	if event == nil {
		return
	}

	if isCancelled(ctx, stopCh) {
		*stopped = true
		return
	}

	if event.Err != nil {
		logx.Errorf("[agent-rpc] ERROR: %v", event.Err)
		_ = stream.Send(&agent.ChatEvent{Type: "error", Content: event.Err.Error()})
		return
	}

	if event.Action != nil {
		if event.Action.Interrupted != nil {
			_ = stream.Send(&agent.ChatEvent{Type: "thought", Content: "等待用户操作..."})
			return
		}
		if event.Action.Exit {
			return
		}
		logx.WithContext(ctx).Debugf("[agent-rpc] unhandled action: exit=%v interrupted=%v",
			event.Action.Exit, event.Action.Interrupted != nil)
	}

	if event.Output != nil && event.Output.MessageOutput != nil {
		mv := event.Output.MessageOutput

		switch mv.Role {
		case schema.Assistant:
			if mv.IsStreaming && mv.MessageStream != nil {
				logx.WithContext(ctx).Debugf("[agent-rpc] assistant streaming...")
				for {
					if isCancelled(ctx, stopCh) {
						*stopped = true
						return
					}
					chunk, err := mv.MessageStream.Recv()
					if err == io.EOF {
						mv.MessageStream.Close()
						break
					}
					if err != nil {
						logx.Errorf("[agent-rpc] stream recv err: %v", err)
						break
					}
					if chunk != nil {
						delta := chunk.Content
						if chunk.ReasoningContent != "" {
							_ = stream.Send(&agent.ChatEvent{Type: "thought", Content: chunk.ReasoningContent})
						}
						if delta != "" {
							*fullReply += delta
							_ = stream.Send(&agent.ChatEvent{Type: "text_delta", Content: delta})
						}
					}
				}
			} else if mv.Message != nil {
				content := mv.Message.Content
				logx.WithContext(ctx).Debugf("[agent-rpc] assistant msg: %s", truncate(content, 100))
				if content != "" {
					*fullReply += content
					_ = stream.Send(&agent.ChatEvent{Type: "text_delta", Content: content})
				}
			} else {
				logx.WithContext(ctx).Debugf("[agent-rpc] assistant: IsStreaming=%v Message=nil", mv.IsStreaming)
			}

		case schema.Tool:
			toolName := mv.ToolName
			logx.WithContext(ctx).Debugf("[agent-rpc] tool: %s streaming=%v", toolName, mv.IsStreaming)
			_ = stream.Send(&agent.ChatEvent{
				Type:     "tool_call",
				ToolName: toolName,
				Content:  "调用工具...",
			})
			content := ""
			if mv.Message != nil {
				content = mv.Message.Content
			}
			_ = stream.Send(&agent.ChatEvent{
				Type:     "tool_result",
				ToolName: toolName,
				ToolData: content,
			})

		default:
			logx.WithContext(ctx).Debugf("[agent-rpc] unknown role: %v", mv.Role)
			if mv.Message != nil && mv.Message.Content != "" {
				*fullReply += mv.Message.Content
				_ = stream.Send(&agent.ChatEvent{Type: "text_delta", Content: mv.Message.Content})
			}
		}
	} else {
		logx.WithContext(ctx).Debugf("[agent-rpc] event has no MessageOutput: output=%v", event.Output != nil)
	}
}

func isCancelled(ctx context.Context, stopCh chan struct{}) bool {
	select {
	case <-stopCh:
		return true
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
