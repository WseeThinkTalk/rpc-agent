package server

import (
	"context"
	"io"
	"sync"
	"time"

	"rpc-agent/agent/internal/agent"
	"rpc-agent/agent/internal/logic"
	"rpc-agent/agent/internal/svc"
	"rpc-agent/agent/internal/tools"
	"rpc-agent/agent/pb"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc"
)

type AgentServer struct {
	pb.UnimplementedAgentServer
	svcCtx    *svc.ServiceContext
	agentMgr  *agent.AgentManager
	sessMgr   *logic.SessionManager
	stopChans map[string]chan struct{}
	mu        sync.Mutex
}

func NewAgentServer(svcCtx *svc.ServiceContext) *AgentServer {
	s := &AgentServer{
		svcCtx:    svcCtx,
		agentMgr:  agent.NewAgentManager(svcCtx.Config.Model),
		sessMgr:   logic.NewSessionManager(svcCtx.Redis, svcCtx.Config.SessionTTL),
		stopChans: make(map[string]chan struct{}),
	}
	if err := s.agentMgr.Init(context.Background(), tools.All(svcCtx)); err != nil {
		panic("failed to init agent: " + err.Error())
	}
	return s
}

func (s *AgentServer) Chat(req *pb.ChatRequest, stream grpc.ServerStreamingServer[pb.ChatEvent]) error {
	ctx := stream.Context()
	logx.Infof("[agent-rpc] Chat called: userId=%d sessionId=%s msg=%s", req.UserId, req.SessionId, req.Message)

	sessionKey := logic.SessionKey(req.UserId, req.SessionId)

	sess, _ := s.sessMgr.Load(ctx, req.UserId, req.SessionId)
	sess.Messages = append(sess.Messages, logic.Message{Role: "user", Content: req.Message})

	// 裁剪上下文
	const maxContextMessages = 20
	ctxMessages := sess.Messages
	if len(ctxMessages) > maxContextMessages {
		ctxMessages = ctxMessages[len(ctxMessages)-maxContextMessages:]
	}

	// 可取消的 context：用户点 Stop 或连接断开时立即中断
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	stopCh := make(chan struct{}, 1)
	s.mu.Lock()
	s.stopChans[sessionKey] = stopCh
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.stopChans, sessionKey)
		s.mu.Unlock()
	}()

	// 监听 stop：关闭 stopCh 或 runCtx 被取消
	go func() {
		select {
		case <-stopCh:
			cancel()
		case <-runCtx.Done():
		}
	}()

	iter := s.agentMgr.Run(runCtx, logic.MessagesToSchema(ctxMessages))
	logx.WithContext(runCtx).Debugf("[agent-rpc] Agent.Run returned iterator, starting event loop")

	// 在 goroutine 中拉取事件，避免 Next() 阻塞主循环
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

	// 心跳 ticker：在等待 LLM 响应时告知前端"还在工作中"
	heartbeat := time.NewTicker(3 * time.Second)
	defer heartbeat.Stop()

	var fullReply string
	stopped := false

	for {
		select {
		case <-heartbeat.C:
			// 心跳：不阻塞，忽略 Send 错误（客户端可能已断开）
			_ = stream.Send(&pb.ChatEvent{Type: "thought", Content: "正在思考..."})

		case <-runCtx.Done():
			_ = stream.Send(&pb.ChatEvent{Type: "done", FinishReason: "stopped"})
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

			s.processEvent(runCtx, stopCh, event, stream, &fullReply, &stopped)
			if stopped {
				_ = stream.Send(&pb.ChatEvent{Type: "done", FinishReason: "stopped"})
				return nil
			}
		}
	}

saveAndDone:
	sess.Messages = append(sess.Messages, logic.Message{Role: "assistant", Content: fullReply})
	s.sessMgr.Save(ctx, req.UserId, req.SessionId, sess)

	stream.Send(&pb.ChatEvent{Type: "done", FinishReason: "stop"})
	return nil
}

func (s *AgentServer) Stop(ctx context.Context, req *pb.StopRequest) (*pb.StopResponse, error) {
	sessionKey := logic.SessionKey(req.UserId, req.SessionId)
	s.mu.Lock()
	defer s.mu.Unlock()
	if ch, ok := s.stopChans[sessionKey]; ok {
		close(ch)
	}
	return &pb.StopResponse{}, nil
}

func (s *AgentServer) ListSessions(ctx context.Context, req *pb.ListSessionsRequest) (*pb.ListSessionsResponse, error) {
	summaries, err := s.sessMgr.ListSessions(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	pbSessions := make([]*pb.SessionSummary, len(summaries))
	for i, sum := range summaries {
		pbSessions[i] = &pb.SessionSummary{
			SessionId:    sum.SessionID,
			Title:        sum.Title,
			MessageCount: sum.MessageCount,
			CreatedAt:    sum.CreatedAt,
			UpdatedAt:    sum.UpdatedAt,
		}
	}
	return &pb.ListSessionsResponse{Sessions: pbSessions}, nil
}

func (s *AgentServer) DeleteSession(ctx context.Context, req *pb.DeleteSessionRequest) (*pb.DeleteSessionResponse, error) {
	if err := s.sessMgr.DeleteSession(ctx, req.UserId, req.SessionId); err != nil {
		return &pb.DeleteSessionResponse{Success: false}, err
	}
	return &pb.DeleteSessionResponse{Success: true}, nil
}

func (s *AgentServer) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	sess, err := s.sessMgr.Load(ctx, req.UserId, req.SessionId)
	if err != nil {
		return nil, err
	}
	msgs := make([]*pb.HistoryMessage, len(sess.Messages))
	for i, m := range sess.Messages {
		msgs[i] = &pb.HistoryMessage{Role: m.Role, Content: m.Content}
	}
	title := sess.Title
	if title == "" {
		for _, m := range sess.Messages {
			if m.Role == "user" {
				title = logic.TruncateRunes(m.Content, 30)
				break
			}
		}
	}
	return &pb.GetHistoryResponse{
		SessionId: req.SessionId,
		Title:     title,
		Messages:  msgs,
		CreatedAt: sess.CreatedAt,
		UpdatedAt: sess.UpdatedAt,
	}, nil
}

func (s *AgentServer) processEvent(ctx context.Context, stopCh chan struct{}, event *adk.AgentEvent, stream grpc.ServerStreamingServer[pb.ChatEvent], fullReply *string, stopped *bool) {
	if event == nil {
		return
	}

	// 上下文已取消 或 用户手动停止 → 立即退出
	if isCancelled(ctx, stopCh) {
		*stopped = true
		return
	}

	// 错误事件
	if event.Err != nil {
		logx.Errorf("[agent-rpc] ERROR: %v", event.Err)
		_ = stream.Send(&pb.ChatEvent{Type: "error", Content: event.Err.Error()})
		return
	}

	// Action 事件
	if event.Action != nil {
		if event.Action.Interrupted != nil {
			_ = stream.Send(&pb.ChatEvent{Type: "thought", Content: "等待用户操作..."})
			return
		}
		if event.Action.Exit {
			return
		}
		logx.WithContext(ctx).Debugf("[agent-rpc] unhandled action: exit=%v interrupted=%v",
			event.Action.Exit, event.Action.Interrupted != nil)
	}

	// Output 事件
	if event.Output != nil && event.Output.MessageOutput != nil {
		mv := event.Output.MessageOutput

		switch mv.Role {
		case schema.Assistant:
			if mv.IsStreaming && mv.MessageStream != nil {
				logx.WithContext(ctx).Debugf("[agent-rpc] assistant streaming...")
				for {
					// 每轮检查上下文和停止信号
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
							_ = stream.Send(&pb.ChatEvent{Type: "thought", Content: chunk.ReasoningContent})
						}
						if delta != "" {
							*fullReply += delta
							_ = stream.Send(&pb.ChatEvent{Type: "text_delta", Content: delta})
						}
					}
				}
			} else if mv.Message != nil {
				content := mv.Message.Content
				logx.WithContext(ctx).Debugf("[agent-rpc] assistant msg: %s", truncate(content, 100))
				if content != "" {
					*fullReply += content
					_ = stream.Send(&pb.ChatEvent{Type: "text_delta", Content: content})
				}
			} else {
				logx.WithContext(ctx).Debugf("[agent-rpc] assistant: IsStreaming=%v Message=nil", mv.IsStreaming)
			}

		case schema.Tool:
			toolName := mv.ToolName
			logx.WithContext(ctx).Debugf("[agent-rpc] tool: %s streaming=%v", toolName, mv.IsStreaming)
			_ = stream.Send(&pb.ChatEvent{
				Type:     "tool_call",
				ToolName: toolName,
				Content:  "调用工具...",
			})
			content := ""
			if mv.Message != nil {
				content = mv.Message.Content
			}
			_ = stream.Send(&pb.ChatEvent{
				Type:     "tool_result",
				ToolName: toolName,
				ToolData: content,
			})

		default:
			logx.WithContext(ctx).Debugf("[agent-rpc] unknown role: %v", mv.Role)
			if mv.Message != nil && mv.Message.Content != "" {
				*fullReply += mv.Message.Content
				_ = stream.Send(&pb.ChatEvent{Type: "text_delta", Content: mv.Message.Content})
			}
		}
	} else {
		logx.WithContext(ctx).Debugf("[agent-rpc] event has no MessageOutput: output=%v", event.Output != nil)
	}
}

// isCancelled 检查上下文是否已取消 或 停止信号是否已触发
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
	if len(s) <= n { return s }
	return s[:n] + "..."
}
