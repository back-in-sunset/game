package domain

// CallState 表示 1v1 通话状态。
type CallState string

const (
	CallStateRinging    CallState = "ringing"    // 呼叫中（等待被叫接听）
	CallStateConnected  CallState = "connected"  // 已接通
	CallStateEnded      CallState = "ended"      // 已挂断
	CallStateRejected   CallState = "rejected"   // 被叫拒绝
	CallStateCancelled  CallState = "cancelled"  // 主叫取消
)

type RoomState string

const (
	RoomStateActive   RoomState = "active"
	RoomStateInactive RoomState = "inactive"
)

// Participant 房间参与者及其静音状态。
type Participant struct {
	UserID int64 `json:"user_id"`
	Muted  bool  `json:"muted"`
}

// VoiceCall 1v1 通话实例（Redis 中存储）。
type VoiceCall struct {
	CallID      string    `json:"call_id"`
	CallerID    int64     `json:"caller_id"`    // 主叫
	CalleeID    int64     `json:"callee_id"`    // 被叫
	Domain      string    `json:"domain"`
	TenantID    string    `json:"tenant_id"`
	ProjectID   string    `json:"project_id"`
	Environment string    `json:"environment"`
	State       CallState `json:"state"`
	LiveKitRoom string    `json:"livekit_room"` // LiveKit 房间名
}

// VoiceRoom 语音房间配置（暂存 Redis）。
type VoiceRoom struct {
	RoomID          string `json:"room_id"`
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	TenantID        string `json:"tenant_id"`
	LiveKitRoom     string `json:"livekit_room"`
	MaxParticipants int    `json:"max_participants"`
}

// CallerInfo 呼叫者身份，从 IM principal 提取后跨服务传递。
type CallerInfo struct {
	UserID      int64
	Domain      string // "platform" | "tenant"
	TenantID    string
	ProjectID   string
	Environment string
}

func (c CallerInfo) PrincipalKey() string {
	switch c.Domain {
	case "platform":
		return "platform:" + itoa(c.UserID)
	case "tenant":
		return "tenant:" + c.TenantID + ":" + c.ProjectID + ":" + c.Environment + ":" + itoa(c.UserID)
	default:
		return "unknown:" + itoa(c.UserID)
	}
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	s := ""
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	if neg {
		s = "-" + s
	}
	return s
}
