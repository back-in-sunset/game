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

type Client struct {
	host      string
	apiKey    string
	apiSecret string
}

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

// GenerateToken creates a LiveKit access token for a user to join a room.
// The token grants both publish (microphone) and subscribe permissions.
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

// RoomNameForCall generates a deterministic room name for a private call.
func RoomNameForCall(callID string) string {
	return "call_" + callID
}

// RoomNameForRoom generates a deterministic room name prefix for a voice room.
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
