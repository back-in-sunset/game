package types

const UserIDKey = "user_id"

type RecordHistoryRequest struct {
	UserID     int64  `json:"user_id,optional" form:"user_id,optional"`
	MediaType  int64  `json:"media_type" form:"media_type"`
	MediaID    int64  `json:"media_id" form:"media_id"`
	Title      string `json:"title,optional" form:"title,optional"`
	Cover      string `json:"cover,optional" form:"cover,optional"`
	AuthorID   int64  `json:"author_id,optional" form:"author_id,optional"`
	ProgressMs int64  `json:"progress_ms,optional" form:"progress_ms,optional"`
	DurationMs int64  `json:"duration_ms,optional" form:"duration_ms,optional"`
	Finished   bool   `json:"finished,optional" form:"finished,optional"`
	Source     int64  `json:"source,optional" form:"source,optional"`
	Device     string `json:"device,optional" form:"device,optional"`
	Meta       string `json:"meta,optional" form:"meta,optional"`
}

type HistoryRecord struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	MediaType   int64  `json:"media_type"`
	MediaID     int64  `json:"media_id"`
	Title       string `json:"title"`
	Cover       string `json:"cover"`
	AuthorID    int64  `json:"author_id"`
	ProgressMs  int64  `json:"progress_ms"`
	DurationMs  int64  `json:"duration_ms"`
	Finished    bool   `json:"finished"`
	Source      int64  `json:"source"`
	Device      string `json:"device"`
	Meta        string `json:"meta"`
	FirstSeenAt int64  `json:"first_seen_at"`
	LastSeenAt  int64  `json:"last_seen_at"`
}

type RecordHistoryResponse struct {
	Record *HistoryRecord `json:"record"`
}

type ListHistoryRequest struct {
	UserID    int64 `form:"user_id,optional"`
	MediaType int64 `form:"media_type,optional"`
	Cursor    int64 `form:"cursor,optional"`
	LastID    int64 `form:"last_id,optional"`
	PageSize  int64 `form:"page_size,optional"`
}

type ListHistoryResponse struct {
	List   []HistoryRecord `json:"list"`
	IsEnd  bool            `json:"is_end"`
	Cursor int64           `json:"cursor"`
	LastID int64           `json:"last_id"`
}

type DeleteHistoryItemRequest struct {
	UserID    int64 `json:"user_id,optional" form:"user_id,optional"`
	MediaType int64 `json:"media_type" form:"media_type"`
	MediaID   int64 `json:"media_id" form:"media_id"`
}

type ClearHistoryByTypeRequest struct {
	UserID    int64 `json:"user_id,optional" form:"user_id,optional"`
	MediaType int64 `json:"media_type" form:"media_type"`
}

type ClearHistoryAllRequest struct {
	UserID int64 `json:"user_id,optional" form:"user_id,optional"`
}

type ActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
