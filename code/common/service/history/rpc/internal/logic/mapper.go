package logic

import (
	"history/model"
	"history/rpc/historyclient"
)

func toRPCRecord(in *model.HistoryRecord) *historyclient.HistoryRecord {
	if in == nil {
		return nil
	}
	var firstSeenAt, lastSeenAt int64
	if in.FirstSeenAt.Valid {
		firstSeenAt = in.FirstSeenAt.Time.Unix()
	}
	if in.LastSeenAt.Valid {
		lastSeenAt = in.LastSeenAt.Time.Unix()
	}
	return &historyclient.HistoryRecord{
		ID:          in.ID,
		UserID:      in.UserID,
		MediaType:   in.MediaType,
		MediaID:     in.MediaID,
		Title:       in.Title,
		Cover:       in.Cover,
		AuthorID:    in.AuthorID,
		ProgressMs:  in.ProgressMs,
		DurationMs:  in.DurationMs,
		Finished:    in.Finished == 1,
		Source:      in.Source,
		Device:      in.Device,
		Meta:        in.Meta,
		FirstSeenAt: firstSeenAt,
		LastSeenAt:  lastSeenAt,
	}
}
