package historyclient

import (
	"context"
	"errors"
	"net/http"

	"history/internal/errx"
	"history/internal/historycache"
	"history/model"

	"google.golang.org/grpc"
)

type localHistory struct {
	manager *historycache.Manager
}

func NewLocalHistory(manager *historycache.Manager) History {
	return &localHistory{manager: manager}
}

func (h *localHistory) RecordHistory(ctx context.Context, in *RecordHistoryRequest, _ ...grpc.CallOption) (*RecordHistoryResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := validateMedia(in.MediaType, in.MediaID); err != nil {
		return nil, err
	}
	finished := int64(0)
	if in.Finished {
		finished = 1
	}
	record, err := h.manager.Record(ctx, &model.HistoryRecord{
		UserID:     in.UserID,
		MediaType:  in.MediaType,
		MediaID:    in.MediaID,
		Title:      in.Title,
		Cover:      in.Cover,
		AuthorID:   in.AuthorID,
		ProgressMs: in.ProgressMs,
		DurationMs: in.DurationMs,
		Finished:   finished,
		Source:     in.Source,
		Device:     in.Device,
		Meta:       in.Meta,
	})
	if err != nil {
		return nil, mapModelError(err)
	}
	return &RecordHistoryResponse{Record: toClientRecord(record)}, nil
}

func (h *localHistory) ListHistory(ctx context.Context, in *ListHistoryRequest, _ ...grpc.CallOption) (*ListHistoryResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if in.MediaType != 0 {
		if err := validateMedia(in.MediaType, 1); err != nil {
			return nil, err
		}
	}
	if in.PageSize < 0 || in.PageSize > model.MaxPageSize {
		return nil, errx.RPCError(http.StatusBadRequest, errx.CodePageSizeInvalid, "page_size invalid")
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = model.DefaultPageSize
	}
	result, err := h.manager.List(ctx, in.UserID, in.MediaType, in.Cursor, in.LastID, pageSize)
	if err != nil {
		return nil, mapModelError(err)
	}
	out := &ListHistoryResponse{
		List:   make([]*HistoryRecord, 0, len(result.Records)),
		IsEnd:  result.IsEnd,
		Cursor: result.Cursor,
		LastID: result.LastID,
	}
	for _, record := range result.Records {
		out.List = append(out.List, toClientRecord(record))
	}
	return out, nil
}

func (h *localHistory) DeleteHistoryItem(ctx context.Context, in *DeleteHistoryItemRequest, _ ...grpc.CallOption) (*ActionResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := validateMedia(in.MediaType, in.MediaID); err != nil {
		return nil, err
	}
	if err := h.manager.DeleteItem(ctx, in.UserID, in.MediaType, in.MediaID); err != nil {
		return nil, mapModelError(err)
	}
	return &ActionResponse{Success: true, Message: "ok"}, nil
}

func (h *localHistory) ClearHistoryByType(ctx context.Context, in *ClearHistoryByTypeRequest, _ ...grpc.CallOption) (*ActionResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := validateMedia(in.MediaType, 1); err != nil {
		return nil, err
	}
	if err := h.manager.ClearByType(ctx, in.UserID, in.MediaType); err != nil {
		return nil, mapModelError(err)
	}
	return &ActionResponse{Success: true, Message: "ok"}, nil
}

func (h *localHistory) ClearHistoryAll(ctx context.Context, in *ClearHistoryAllRequest, _ ...grpc.CallOption) (*ActionResponse, error) {
	if err := validateUserID(in.UserID); err != nil {
		return nil, err
	}
	if err := h.manager.ClearAll(ctx, in.UserID); err != nil {
		return nil, mapModelError(err)
	}
	return &ActionResponse{Success: true, Message: "ok"}, nil
}

func validateUserID(userID int64) error {
	if userID <= 0 {
		return errx.RPCError(http.StatusBadRequest, errx.CodeUserIDInvalid, "user_id invalid")
	}
	return nil
}

func validateMedia(mediaType int64, mediaID int64) error {
	if mediaType != model.MediaTypePost && mediaType != model.MediaTypeVideo {
		return errx.RPCError(http.StatusBadRequest, errx.CodeMediaTypeInvalid, "media_type invalid")
	}
	if mediaID <= 0 {
		return errx.RPCError(http.StatusBadRequest, errx.CodeMediaIDInvalid, "media_id invalid")
	}
	return nil
}

func mapModelError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, model.ErrInvalidProgress) {
		return errx.RPCError(http.StatusBadRequest, errx.CodeProgressInvalid, "progress invalid")
	}
	return errx.RPCError(http.StatusInternalServerError, errx.CodeInternalDefault, err.Error())
}

func toClientRecord(in *model.HistoryRecord) *HistoryRecord {
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
	return &HistoryRecord{
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
