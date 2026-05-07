package livekit

import (
	"fmt"
	"strconv"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
)

const (
	tokenTTL       = 24 * time.Hour
	refreshTTL     = 48 * time.Hour
	roomCreateTTL  = 10 * time.Minute
	roomEmptyLimit = 5 * time.Minute
)

// Client 封装 LiveKit 服务器 API，负责 token 签发。
type Client struct {
	host      string // LiveKit 服务地址
	apiKey    string
	apiSecret string
}

// New 创建 LiveKit 客户端。
func New(host, apiKey, apiSecret string) *Client {
	return &Client{
		host:      host,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

func (c *Client) Host() string {
	return c.host
}

// GenerateToken 签发 LiveKit 访问 JWT，授予麦克风发布和订阅权限。
func (c *Client) GenerateToken(roomName string, userID int64, userName string) (string, error) {
	identity := userName
	if identity == "" {
		identity = strconv.FormatInt(userID, 10)
	}

	at := auth.NewAccessToken(c.apiKey, c.apiSecret)
	at.SetName(identity)
	at.SetIdentity(strconv.FormatInt(userID, 10))
	at.SetValidFor(tokenTTL)

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
		Hidden:   false,
	}
	grant.SetCanPublish(true)
	grant.SetCanPublishData(true)
	grant.SetCanSubscribe(true)
	at.SetVideoGrant(grant)

	return at.ToJWT()
}

// RoomNameForCall 生成 1v1 通话的 LiveKit 房间名（前缀 call_）。
func RoomNameForCall(callID string) string {
	return "call_" + callID
}

// RoomNameForRoom 生成语音房间的 LiveKit 房间名（前缀 room_）。
func RoomNameForRoom(roomID string) string {
	return "room_" + roomID
}

// stub returns a stub config for room auto-creation.
func (c *Client) roomConfig(name string) *livekit.RoomConfiguration {
	return &livekit.RoomConfiguration{
		Name:             name,
		MaxParticipants:  25,
		EmptyTimeout:     uint32(roomEmptyLimit.Seconds()),
		DepartureTimeout: 10,
		MinPlayoutDelay:  0,
		MaxPlayoutDelay:  0,
	}
}

func (c *Client) String() string {
	return fmt.Sprintf("livekit client host=%s", c.host)
}
