package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type SessionManager struct {
	redis *redis.Redis
	ttl   time.Duration
}

type Session struct {
	Title     string    `json:"title"`
	Messages  []Message `json:"messages"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

type Message struct {
	Role       string `json:"role"`
	Content    string `json:"content"`
	ToolName   string `json:"tool_name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
}

type SessionSummary struct {
	SessionID    string `json:"session_id"`
	Title        string `json:"title"`
	MessageCount int32  `json:"message_count"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

func NewSessionManager(rds *redis.Redis, ttl int64) *SessionManager {
	return &SessionManager{redis: rds, ttl: time.Duration(ttl) * time.Second}
}

func SessionKey(userID int64, sessionID string) string {
	return fmt.Sprintf("%d:%s", userID, sessionID)
}

func (sm *SessionManager) key(userID int64, sessionID string) string {
	return fmt.Sprintf("agent:session:%d:%s", userID, sessionID)
}

func (sm *SessionManager) sessionsIndexKey(userID int64) string {
	return fmt.Sprintf("agent:sessions:%d", userID)
}

func (sm *SessionManager) Load(ctx context.Context, userID int64, sessionID string) (*Session, error) {
	data, err := sm.redis.GetCtx(ctx, sm.key(userID, sessionID))
	if err != nil || data == "" {
		return &Session{
			Messages:  []Message{},
			CreatedAt: time.Now().Unix(),
			UpdatedAt: time.Now().Unix(),
		}, nil
	}
	var sess Session
	if err := json.Unmarshal([]byte(data), &sess); err != nil {
		return &Session{Messages: []Message{}, CreatedAt: time.Now().Unix()}, nil
	}
	return &sess, nil
}

func (sm *SessionManager) Save(ctx context.Context, userID int64, sessionID string, sess *Session) error {
	sess.UpdatedAt = time.Now().Unix()
	if len(sess.Messages) > 100 {
		sess.Messages = sess.Messages[len(sess.Messages)-100:]
	}

	// 自动标题：取用户第一条消息截断 30 个 rune
	if sess.Title == "" && len(sess.Messages) > 0 {
		for _, m := range sess.Messages {
			if m.Role == "user" {
				sess.Title = TruncateRunes(m.Content, 30)
				break
			}
		}
	}

	data, _ := json.Marshal(sess)
	if err := sm.redis.SetexCtx(ctx, sm.key(userID, sessionID), string(data), int(sm.ttl.Seconds())); err != nil {
		return err
	}
	// 更新 sorted set 索引
	if _, err := sm.redis.ZaddCtx(ctx, sm.sessionsIndexKey(userID), sess.UpdatedAt, sessionID); err != nil {
		return err
	}
	sm.redis.ExpireCtx(ctx, sm.sessionsIndexKey(userID), int(sm.ttl.Seconds()))
	return nil
}

func (sm *SessionManager) ListSessions(ctx context.Context, userID int64) ([]SessionSummary, error) {
	members, err := sm.redis.ZrevrangeWithScoresCtx(ctx, sm.sessionsIndexKey(userID), 0, 49)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return []SessionSummary{}, nil
	}

	// 批量 GET，一次网络往返
	keys := make([]string, len(members))
	for i, m := range members {
		keys[i] = sm.key(userID, m.Key)
	}
	values, err := sm.redis.MgetCtx(ctx, keys...)
	if err != nil {
		return nil, err
	}

	summaries := make([]SessionSummary, 0, len(members))
	for i, data := range values {
		if data == "" {
			continue
		}
		var sess Session
		if err := json.Unmarshal([]byte(data), &sess); err != nil {
			continue
		}
		title := sess.Title
		if title == "" {
			for _, msg := range sess.Messages {
				if msg.Role == "user" {
					title = TruncateRunes(msg.Content, 30)
					break
				}
			}
		}
		summaries = append(summaries, SessionSummary{
			SessionID:    members[i].Key,
			Title:        title,
			MessageCount: int32(len(sess.Messages)),
			CreatedAt:    sess.CreatedAt,
			UpdatedAt:    sess.UpdatedAt,
		})
	}
	return summaries, nil
}

func (sm *SessionManager) DeleteSession(ctx context.Context, userID int64, sessionID string) error {
	if _, err := sm.redis.DelCtx(ctx, sm.key(userID, sessionID)); err != nil {
		return err
	}
	if _, err := sm.redis.ZremCtx(ctx, sm.sessionsIndexKey(userID), sessionID); err != nil {
		return err
	}
	return nil
}

func TruncateRunes(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func MessagesToSchema(msgs []Message) []*schema.Message {
	result := make([]*schema.Message, len(msgs))
	for i, m := range msgs {
		result[i] = &schema.Message{
			Role:    schema.RoleType(m.Role),
			Content: m.Content,
		}
	}
	return result
}
