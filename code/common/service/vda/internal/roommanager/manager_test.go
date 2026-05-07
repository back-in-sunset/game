package roommanager

import (
	"context"
	"slices"
	"strings"
	"testing"

	"vda/internal/domain"
	"vda/internal/storage"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type stubTokenGen struct{}

func (s *stubTokenGen) GenerateToken(roomName string, userID int64, userName string) (string, error) {
	return "fake-token-" + roomName, nil
}

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}
	t.Cleanup(mr.Close)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })

	// Use prefix "vda" because Manager.GetMuted hardcodes storePrefix() to "vda".
	store := storage.NewRedisStore(rdb, "vda")
	return New(&stubTokenGen{}, store, rdb)
}

func callerInfo(userID int64) domain.CallerInfo {
	return domain.CallerInfo{
		UserID:      userID,
		Domain:      "platform",
		TenantID:    "",
		ProjectID:   "",
		Environment: "",
	}
}

func TestManager_Join_Success(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	token, participants, err := mgr.Join(ctx, "room-1", callerInfo(1001))
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if token == "" {
		t.Fatal("Join() token is empty")
	}
	if !strings.Contains(token, "fake-token-") {
		t.Fatalf("Join() token = %q, want containing 'fake-token-'", token)
	}
	if !slices.Contains(participants, int64(1001)) {
		t.Fatalf("Join() participants = %v, want containing 1001", participants)
	}
}

func TestManager_Join_MultipleUsers(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join(1001) error = %v", err)
	}
	_, participants, err := mgr.Join(ctx, "room-1", callerInfo(2002))
	if err != nil {
		t.Fatalf("Join(2002) error = %v", err)
	}

	has1, has2 := false, false
	for _, id := range participants {
		if id == 1001 {
			has1 = true
		}
		if id == 2002 {
			has2 = true
		}
	}
	if !has1 || !has2 {
		t.Fatalf("Join() participants = %v, want containing 1001 and 2002", participants)
	}
}

func TestManager_Join_SetsPresence(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	_, _, err := mgr.Join(ctx, "room-1", callerInfo(1001))
	if err != nil {
		t.Fatalf("Join() error = %v", err)
	}

	activity, err := mgr.store.GetVoicePresence(ctx, 1001)
	if err != nil {
		t.Fatalf("GetVoicePresence() error = %v", err)
	}
	if !strings.Contains(activity, "room:") {
		t.Fatalf("GetVoicePresence() = %q, want containing 'room:'", activity)
	}
}

func TestManager_Leave_Success(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if err := mgr.Leave(ctx, "room-1", 1001); err != nil {
		t.Fatalf("Leave() error = %v", err)
	}

	participants, err := mgr.GetParticipants(ctx, "room-1")
	if err != nil {
		t.Fatalf("GetParticipants() error = %v", err)
	}
	for _, p := range participants {
		if p.UserID == 1001 {
			t.Fatal("GetParticipants() contains 1001 after leave")
		}
	}
}

func TestManager_Leave_ClearsPresence(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if err := mgr.Leave(ctx, "room-1", 1001); err != nil {
		t.Fatalf("Leave() error = %v", err)
	}

	activity, err := mgr.store.GetVoicePresence(ctx, 1001)
	if err != nil {
		t.Fatalf("GetVoicePresence() error = %v", err)
	}
	if activity != "" {
		t.Fatalf("GetVoicePresence() after leave = %q, want empty", activity)
	}
}

func TestManager_ToggleMute_Mute(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if err := mgr.ToggleMute(ctx, "room-1", 1001, true); err != nil {
		t.Fatalf("ToggleMute(true) error = %v", err)
	}

	muted, err := mgr.store.IsMuted(ctx, "room-1", 1001)
	if err != nil {
		t.Fatalf("IsMuted() error = %v", err)
	}
	if !muted {
		t.Fatal("IsMuted() = false, want true")
	}
}

func TestManager_ToggleMute_Unmute(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if err := mgr.ToggleMute(ctx, "room-1", 1001, true); err != nil {
		t.Fatalf("ToggleMute(true) error = %v", err)
	}
	if err := mgr.ToggleMute(ctx, "room-1", 1001, false); err != nil {
		t.Fatalf("ToggleMute(false) error = %v", err)
	}

	muted, err := mgr.store.IsMuted(ctx, "room-1", 1001)
	if err != nil {
		t.Fatalf("IsMuted() error = %v", err)
	}
	if muted {
		t.Fatal("IsMuted() = true, want false after unmute")
	}
}

func TestManager_GetParticipants_WithMuteStates(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join(1001) error = %v", err)
	}
	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(2002)); err != nil {
		t.Fatalf("Join(2002) error = %v", err)
	}
	if err := mgr.ToggleMute(ctx, "room-1", 1001, true); err != nil {
		t.Fatalf("ToggleMute() error = %v", err)
	}

	participants, err := mgr.GetParticipants(ctx, "room-1")
	if err != nil {
		t.Fatalf("GetParticipants() error = %v", err)
	}

	for _, p := range participants {
		switch p.UserID {
		case 1001:
			if !p.Muted {
				t.Fatal("participant 1001 Muted = false, want true")
			}
		case 2002:
			if p.Muted {
				t.Fatal("participant 2002 Muted = true, want false")
			}
		}
	}
}

func TestManager_GetMuted(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join(1001) error = %v", err)
	}
	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(2002)); err != nil {
		t.Fatalf("Join(2002) error = %v", err)
	}
	if err := mgr.ToggleMute(ctx, "room-1", 1001, true); err != nil {
		t.Fatalf("ToggleMute(1001) error = %v", err)
	}
	if err := mgr.ToggleMute(ctx, "room-1", 2002, true); err != nil {
		t.Fatalf("ToggleMute(2002) error = %v", err)
	}

	muted, err := mgr.GetMuted(ctx, "room-1")
	if err != nil {
		t.Fatalf("GetMuted() error = %v", err)
	}

	has1, has2 := false, false
	for _, id := range muted {
		if id == 1001 {
			has1 = true
		}
		if id == 2002 {
			has2 = true
		}
	}
	if !has1 || !has2 {
		t.Fatalf("GetMuted() = %v, want containing 1001 and 2002", muted)
	}
}

func TestManager_LeaveRemovesMuteState(t *testing.T) {
	mgr := newTestManager(t)
	ctx := context.Background()

	if _, _, err := mgr.Join(ctx, "room-1", callerInfo(1001)); err != nil {
		t.Fatalf("Join() error = %v", err)
	}
	if err := mgr.ToggleMute(ctx, "room-1", 1001, true); err != nil {
		t.Fatalf("ToggleMute() error = %v", err)
	}
	if err := mgr.Leave(ctx, "room-1", 1001); err != nil {
		t.Fatalf("Leave() error = %v", err)
	}

	muted, err := mgr.GetMuted(ctx, "room-1")
	if err != nil {
		t.Fatalf("GetMuted() error = %v", err)
	}
	for _, id := range muted {
		if id == 1001 {
			t.Fatal("GetMuted() contains 1001 after leave")
		}
	}
}
