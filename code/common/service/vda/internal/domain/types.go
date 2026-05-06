package domain

type CallState string

const (
	CallStateRinging    CallState = "ringing"
	CallStateConnected  CallState = "connected"
	CallStateEnded      CallState = "ended"
	CallStateRejected   CallState = "rejected"
	CallStateCancelled  CallState = "cancelled"
)

type RoomState string

const (
	RoomStateActive   RoomState = "active"
	RoomStateInactive RoomState = "inactive"
)

type Participant struct {
	UserID int64  `json:"user_id"`
	Muted  bool   `json:"muted"`
}

type VoiceCall struct {
	CallID      string    `json:"call_id"`
	CallerID    int64     `json:"caller_id"`
	CalleeID    int64     `json:"callee_id"`
	Domain      string    `json:"domain"`
	TenantID    string    `json:"tenant_id"`
	ProjectID   string    `json:"project_id"`
	Environment string    `json:"environment"`
	State       CallState `json:"state"`
	LiveKitRoom string    `json:"livekit_room"`
}

type VoiceRoom struct {
	RoomID       string `json:"room_id"`
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	TenantID     string `json:"tenant_id"`
	LiveKitRoom  string `json:"livekit_room"`
	MaxParticipants int `json:"max_participants"`
}

type CallerInfo struct {
	UserID      int64
	Domain      string
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
