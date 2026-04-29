package logic

import (
	"context"
	"encoding/json"
	"net/http"

	"history/api/internal/types"
	"history/internal/errx"
	"history/rpc/historyclient"
)

func extractUserID(ctx context.Context, fallback int64) (int64, error) {
	if fallback > 0 {
		return fallback, nil
	}
	v := ctx.Value(types.UserIDKey)
	switch x := v.(type) {
	case json.Number:
		uid, err := x.Int64()
		if err == nil && uid > 0 {
			return uid, nil
		}
	case int64:
		if x > 0 {
			return x, nil
		}
	case float64:
		if x > 0 {
			return int64(x), nil
		}
	}
	return 0, errx.New(http.StatusBadRequest, errx.CodeUserIDInvalid, "user_id invalid")
}

func toAPIRecord(in *historyclient.HistoryRecord) *types.HistoryRecord {
	if in == nil {
		return nil
	}
	return &types.HistoryRecord{
		ID:          in.ID,
		UserID:      in.UserID,
		MediaType:   in.MediaType,
		MediaID:     in.MediaID,
		Title:       in.Title,
		Cover:       in.Cover,
		AuthorID:    in.AuthorID,
		ProgressMs:  in.ProgressMs,
		DurationMs:  in.DurationMs,
		Finished:    in.Finished,
		Source:      in.Source,
		Device:      in.Device,
		Meta:        in.Meta,
		FirstSeenAt: in.FirstSeenAt,
		LastSeenAt:  in.LastSeenAt,
	}
}
