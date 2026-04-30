package historycache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"history/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	memberSeparator      = ":"
	deleteMarkerAll      = "all"
	deleteMarkerTypePref = "type:"
	deleteMarkerItemPref = "item:"
	flushLockTTLSeconds  = 30
)

type Config struct {
	ListTTLSeconds       int
	DetailTTLSeconds     int
	FlushIntervalSeconds int
	FlushBatchUsers      int
	FlushBatchItems      int
	DirtyTTLSeconds      int
	DeleteMarkerTTL      int
	ReadFallbackToDB     bool
	WriteBackEnabled     bool
}

type ListResult struct {
	Records []*model.HistoryRecord
	IsEnd   bool
	Cursor  int64
	LastID  int64
}

type Manager struct {
	model model.HistoryModel
	rds   *redis.Redis
	cfg   Config

	bgOnce sync.Once
	stopFn context.CancelFunc
}

type cachedRecord struct {
	Record   *model.HistoryRecord
	SortID   int64
	Member   string
	Identity string
}

func NewManager(historyModel model.HistoryModel, rds *redis.Redis, cfg Config) *Manager {
	cfg = cfg.withDefaults()
	return &Manager{
		model: historyModel,
		rds:   rds,
		cfg:   cfg,
	}
}

func (m *Manager) Start() {
	if !m.writeBackEnabled() {
		return
	}
	m.bgOnce.Do(func() {
		bgCtx, cancel := context.WithCancel(context.Background())
		m.stopFn = cancel
		go m.runFlushLoop(bgCtx)
	})
}

func (m *Manager) Stop() {
	if m.stopFn != nil {
		m.stopFn()
	}
}

func (m *Manager) Record(ctx context.Context, data *model.HistoryRecord) (*model.HistoryRecord, error) {
	if !m.cacheEnabled() {
		return m.model.UpsertRecord(ctx, data)
	}

	now := time.Now()
	existing, err := m.loadCachedRecord(ctx, data.UserID, data.MediaType, data.MediaID)
	if err != nil {
		logx.WithContext(ctx).Errorf("load cached history error: %v", err)
		return m.fallbackRecord(ctx, data)
	}

	sortID, err := m.rds.IncrCtx(ctx, m.userSeqKey(data.UserID))
	if err != nil {
		logx.WithContext(ctx).Errorf("alloc history sort id error: %v", err)
		return m.fallbackRecord(ctx, data)
	}

	firstSeenAt := now
	dbID := int64(0)
	oldMember := ""
	if existing != nil && existing.Record != nil {
		oldMember = existing.Member
		dbID = existing.Record.ID
		if existing.Record.FirstSeenAt.Valid {
			firstSeenAt = existing.Record.FirstSeenAt.Time
		}
	}

	record := cloneHistoryRecord(data)
	record.ID = dbID
	record.FirstSeenAt = sql.NullTime{Time: firstSeenAt, Valid: true}
	record.LastSeenAt = sql.NullTime{Time: now, Valid: true}

	member := historyMember(sortID, data.MediaType, data.MediaID)
	if err := m.writeRecordToCache(ctx, record, sortID, member, oldMember); err != nil {
		logx.WithContext(ctx).Errorf("write history cache error: %v", err)
		return m.fallbackRecord(ctx, data)
	}

	return cloneHistoryRecord(record), nil
}

func (m *Manager) List(ctx context.Context, userID, mediaType, cursor, lastID, pageSize int64) (*ListResult, error) {
	if !m.cacheEnabled() {
		return m.listFromDB(ctx, userID, mediaType, cursor, lastID, pageSize)
	}

	result, hit, err := m.listFromCache(ctx, userID, mediaType, cursor, lastID, pageSize)
	if err != nil {
		logx.WithContext(ctx).Errorf("list history cache error: %v", err)
		if m.cfg.ReadFallbackToDB {
			return m.listFromDB(ctx, userID, mediaType, cursor, lastID, pageSize)
		}
		return nil, err
	}
	if hit {
		return result, nil
	}
	if !m.cfg.ReadFallbackToDB {
		return &ListResult{Records: make([]*model.HistoryRecord, 0), IsEnd: true}, nil
	}
	return m.listFromDB(ctx, userID, mediaType, cursor, lastID, pageSize)
}

func (m *Manager) DeleteItem(ctx context.Context, userID, mediaType, mediaID int64) error {
	if !m.cacheEnabled() {
		return m.model.SoftDeleteItem(ctx, userID, mediaType, mediaID)
	}

	existing, err := m.loadCachedRecord(ctx, userID, mediaType, mediaID)
	if err != nil {
		logx.WithContext(ctx).Errorf("load cached history before delete error: %v", err)
		return m.model.SoftDeleteItem(ctx, userID, mediaType, mediaID)
	}

	itemKey := m.itemKey(userID, mediaType, mediaID)
	if _, err := m.rds.DelCtx(ctx, itemKey); err != nil {
		logx.WithContext(ctx).Errorf("delete history cache item error: %v", err)
		return m.model.SoftDeleteItem(ctx, userID, mediaType, mediaID)
	}
	if existing != nil && existing.Member != "" {
		_, _ = m.rds.ZremCtx(ctx, m.userListKey(userID), existing.Member)
		_, _ = m.rds.ZremCtx(ctx, m.userTypeListKey(userID, mediaType), existing.Member)
	}
	if err := m.markDeleteItem(ctx, userID, mediaType, mediaID); err != nil {
		logx.WithContext(ctx).Errorf("mark history delete item error: %v", err)
		return m.model.SoftDeleteItem(ctx, userID, mediaType, mediaID)
	}
	return nil
}

func (m *Manager) ClearByType(ctx context.Context, userID, mediaType int64) error {
	if !m.cacheEnabled() {
		return m.model.SoftDeleteByType(ctx, userID, mediaType)
	}

	records, err := m.membersForListKey(ctx, m.userTypeListKey(userID, mediaType))
	if err != nil {
		logx.WithContext(ctx).Errorf("load history type members error: %v", err)
		return m.model.SoftDeleteByType(ctx, userID, mediaType)
	}

	for _, member := range records {
		sortID, mt, mid, parseErr := parseHistoryMember(member)
		if parseErr != nil {
			continue
		}
		_, _ = m.rds.ZremCtx(ctx, m.userListKey(userID), member)
		_, _ = m.rds.ZremCtx(ctx, m.userTypeListKey(userID, mediaType), historyMember(sortID, mt, mid))
		_, _ = m.rds.DelCtx(ctx, m.itemKey(userID, mt, mid))
		_, _ = m.rds.SremCtx(ctx, m.dirtyItemsKey(userID), historyIdentity(mt, mid))
	}

	_, _ = m.rds.DelCtx(ctx, m.userTypeListKey(userID, mediaType))
	if err := m.markDeleteType(ctx, userID, mediaType); err != nil {
		logx.WithContext(ctx).Errorf("mark history delete type error: %v", err)
		return m.model.SoftDeleteByType(ctx, userID, mediaType)
	}
	return nil
}

func (m *Manager) ClearAll(ctx context.Context, userID int64) error {
	if !m.cacheEnabled() {
		return m.model.SoftDeleteAll(ctx, userID)
	}

	allMembers, err := m.membersForListKey(ctx, m.userListKey(userID))
	if err != nil {
		logx.WithContext(ctx).Errorf("load all history members error: %v", err)
		return m.model.SoftDeleteAll(ctx, userID)
	}
	for _, member := range allMembers {
		_, mt, mid, parseErr := parseHistoryMember(member)
		if parseErr != nil {
			continue
		}
		_, _ = m.rds.DelCtx(ctx, m.itemKey(userID, mt, mid))
	}
	_, _ = m.rds.DelCtx(ctx,
		m.userListKey(userID),
		m.userTypeListKey(userID, model.MediaTypePost),
		m.userTypeListKey(userID, model.MediaTypeVideo),
		m.dirtyItemsKey(userID),
	)
	if err := m.markDeleteAll(ctx, userID); err != nil {
		logx.WithContext(ctx).Errorf("mark history delete all error: %v", err)
		return m.model.SoftDeleteAll(ctx, userID)
	}
	return nil
}

func (m *Manager) FlushOnce(ctx context.Context) error {
	if !m.writeBackEnabled() {
		return nil
	}
	users, err := m.rds.ZrangeCtx(ctx, dirtyUsersKey, 0, int64(m.cfg.FlushBatchUsers-1))
	if err != nil {
		return err
	}
	for _, rawUserID := range users {
		userID, parseErr := strconv.ParseInt(rawUserID, 10, 64)
		if parseErr != nil {
			continue
		}
		locked, lockErr := m.rds.SetnxExCtx(ctx, m.flushLockKey(userID), "1", flushLockTTLSeconds)
		if lockErr != nil || !locked {
			continue
		}
		if flushErr := m.flushUser(ctx, userID); flushErr != nil {
			logx.WithContext(ctx).Errorf("flush history user %d error: %v", userID, flushErr)
		}
		_, _ = m.rds.DelCtx(ctx, m.flushLockKey(userID))
	}
	return nil
}

func (m *Manager) runFlushLoop(ctx context.Context) {
	interval := time.Duration(m.cfg.FlushIntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := m.FlushOnce(ctx); err != nil {
				logx.WithContext(ctx).Errorf("history flush tick error: %v", err)
			}
		}
	}
}

func (m *Manager) flushUser(ctx context.Context, userID int64) error {
	markers, err := m.rds.SmembersCtx(ctx, m.deleteMarkersKey(userID))
	if err != nil {
		return err
	}
	hasAllMarker := false
	typeMarkers := make([]int64, 0, len(markers))
	for _, marker := range markers {
		if marker == deleteMarkerAll {
			hasAllMarker = true
			break
		}
		if strings.HasPrefix(marker, deleteMarkerTypePref) {
			mediaType, parseErr := strconv.ParseInt(strings.TrimPrefix(marker, deleteMarkerTypePref), 10, 64)
			if parseErr == nil {
				typeMarkers = append(typeMarkers, mediaType)
			}
		}
	}

	if hasAllMarker {
		if err := m.model.SoftDeleteAll(ctx, userID); err != nil {
			return err
		}
		_, _ = m.rds.DelCtx(ctx, m.dirtyItemsKey(userID), m.deleteMarkersKey(userID))
		_, _ = m.rds.ZremCtx(ctx, dirtyUsersKey, strconv.FormatInt(userID, 10))
		return nil
	}

	for _, mediaType := range typeMarkers {
		if err := m.model.SoftDeleteByType(ctx, userID, mediaType); err != nil {
			return err
		}
		_, _ = m.rds.SremCtx(ctx, m.deleteMarkersKey(userID), deleteMarkerTypePref+strconv.FormatInt(mediaType, 10))
	}

	items, err := m.rds.SmembersCtx(ctx, m.dirtyItemsKey(userID))
	if err != nil {
		return err
	}

	limit := m.cfg.FlushBatchItems
	if limit > len(items) {
		limit = len(items)
	}
	deleteKey := m.deleteMarkersKey(userID)
	for _, identity := range items[:limit] {
		mediaType, mediaID, parseErr := parseIdentity(identity)
		if parseErr != nil {
			_, _ = m.rds.SremCtx(ctx, m.dirtyItemsKey(userID), identity)
			continue
		}

		cached, cacheErr := m.loadCachedRecord(ctx, userID, mediaType, mediaID)
		if cacheErr != nil {
			return cacheErr
		}

		if cached != nil && cached.Record != nil {
			flushed, flushErr := m.model.UpsertRecord(ctx, cached.Record)
			if flushErr != nil {
				return flushErr
			}
			if flushed != nil {
				_ = m.rds.HsetCtx(ctx, m.itemKey(userID, mediaType, mediaID), "id", strconv.FormatInt(flushed.ID, 10))
			}
			_, _ = m.rds.SremCtx(ctx, deleteKey, deleteItemMarker(mediaType, mediaID))
		} else {
			marked, markErr := m.rds.SismemberCtx(ctx, deleteKey, deleteItemMarker(mediaType, mediaID))
			if markErr != nil {
				return markErr
			}
			if marked {
				if err := m.model.SoftDeleteItem(ctx, userID, mediaType, mediaID); err != nil {
					return err
				}
				_, _ = m.rds.SremCtx(ctx, deleteKey, deleteItemMarker(mediaType, mediaID))
			}
		}

		_, _ = m.rds.SremCtx(ctx, m.dirtyItemsKey(userID), identity)
	}

	remainingDirty, _ := m.rds.ScardCtx(ctx, m.dirtyItemsKey(userID))
	remainingMarkers, _ := m.rds.ScardCtx(ctx, deleteKey)
	if remainingDirty == 0 && remainingMarkers == 0 {
		_, _ = m.rds.ZremCtx(ctx, dirtyUsersKey, strconv.FormatInt(userID, 10))
		_, _ = m.rds.DelCtx(ctx, m.dirtyItemsKey(userID), deleteKey)
	}

	return nil
}

func (m *Manager) fallbackRecord(ctx context.Context, data *model.HistoryRecord) (*model.HistoryRecord, error) {
	if !m.cfg.ReadFallbackToDB && !m.cfg.WriteBackEnabled {
		return nil, errors.New("history cache unavailable")
	}
	return m.model.UpsertRecord(ctx, data)
}

func (m *Manager) listFromDB(ctx context.Context, userID, mediaType, cursor, lastID, pageSize int64) (*ListResult, error) {
	records, err := m.model.ListByUser(ctx, userID, mediaType, cursor, lastID, pageSize+1)
	if err != nil {
		return nil, err
	}
	isEnd := true
	if len(records) > int(pageSize) {
		isEnd = false
		records = records[:pageSize]
	}

	result := &ListResult{
		Records: records,
		IsEnd:   isEnd,
	}
	if len(records) > 0 {
		last := records[len(records)-1]
		if last.LastSeenAt.Valid {
			result.Cursor = last.LastSeenAt.Time.Unix()
		}
		result.LastID = last.ID
	}
	return result, nil
}

func (m *Manager) listFromCache(ctx context.Context, userID, mediaType, cursor, lastID, pageSize int64) (*ListResult, bool, error) {
	listKey := m.userListKey(userID)
	if mediaType > 0 {
		listKey = m.userTypeListKey(userID, mediaType)
	}

	exists, err := m.rds.ExistsCtx(ctx, listKey)
	if err != nil {
		return nil, false, err
	}
	if !exists {
		lockedOut, tombErr := m.hasDeleteMarkers(ctx, userID)
		if tombErr != nil {
			return nil, false, tombErr
		}
		if lockedOut {
			return &ListResult{Records: make([]*model.HistoryRecord, 0), IsEnd: true}, true, nil
		}
		return nil, false, nil
	}

	items, err := m.loadPageFromCache(ctx, userID, mediaType, listKey, cursor, lastID, pageSize+1)
	if err != nil {
		return nil, true, err
	}

	isEnd := true
	if len(items) > int(pageSize) {
		isEnd = false
		items = items[:pageSize]
	}

	result := &ListResult{
		Records: make([]*model.HistoryRecord, 0, len(items)),
		IsEnd:   isEnd,
	}
	for _, item := range items {
		result.Records = append(result.Records, cloneHistoryRecord(item.Record))
	}
	if len(items) > 0 {
		last := items[len(items)-1]
		if last.Record.LastSeenAt.Valid {
			result.Cursor = last.Record.LastSeenAt.Time.Unix()
		}
		result.LastID = last.SortID
	}

	return result, true, nil
}

func (m *Manager) loadPageFromCache(ctx context.Context, userID, mediaType int64, listKey string, cursor, lastID, pageSize int64) ([]*cachedRecord, error) {
	const batchSize = 64

	var (
		offset int
		items  []*cachedRecord
	)

	for len(items) < int(pageSize) {
		var (
			pairs []redis.Pair
			err   error
		)
		if cursor > 0 {
			pairs, err = m.rds.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, listKey, cursor, 0, offset, batchSize)
		} else {
			pairs, err = m.rds.ZrevrangeWithScoresCtx(ctx, listKey, int64(offset), int64(offset+batchSize-1))
		}
		if err != nil {
			return nil, err
		}
		if len(pairs) == 0 {
			break
		}
		offset += len(pairs)

		for _, pair := range pairs {
			sortID, mt, mid, parseErr := parseHistoryMember(pair.Key)
			if parseErr != nil {
				_, _ = m.rds.ZremCtx(ctx, listKey, pair.Key)
				continue
			}
			if cursor > 0 && pair.Score == cursor && sortID >= lastID {
				continue
			}
			if mediaType > 0 && mt != mediaType {
				continue
			}

			cached, loadErr := m.loadCachedRecord(ctx, userID, mt, mid)
			if loadErr != nil {
				return nil, loadErr
			}
			if cached == nil || cached.Record == nil || cached.SortID != sortID {
				_, _ = m.rds.ZremCtx(ctx, m.userListKey(userID), pair.Key)
				_, _ = m.rds.ZremCtx(ctx, m.userTypeListKey(userID, mt), pair.Key)
				continue
			}
			items = append(items, cached)
			if len(items) >= int(pageSize) {
				break
			}
		}

		if len(pairs) < batchSize {
			break
		}
	}

	return items, nil
}

func (m *Manager) loadCachedRecord(ctx context.Context, userID, mediaType, mediaID int64) (*cachedRecord, error) {
	fields, err := m.rds.HgetallCtx(ctx, m.itemKey(userID, mediaType, mediaID))
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, nil
	}

	sortID, err := strconv.ParseInt(fields["sort_id"], 10, 64)
	if err != nil {
		return nil, err
	}
	record, err := historyRecordFromFields(fields)
	if err != nil {
		return nil, err
	}
	return &cachedRecord{
		Record:   record,
		SortID:   sortID,
		Member:   historyMember(sortID, mediaType, mediaID),
		Identity: historyIdentity(mediaType, mediaID),
	}, nil
}

func (m *Manager) writeRecordToCache(ctx context.Context, record *model.HistoryRecord, sortID int64, member, oldMember string) error {
	fields := map[string]string{
		"id":            strconv.FormatInt(record.ID, 10),
		"user_id":       strconv.FormatInt(record.UserID, 10),
		"media_type":    strconv.FormatInt(record.MediaType, 10),
		"media_id":      strconv.FormatInt(record.MediaID, 10),
		"title":         record.Title,
		"cover":         record.Cover,
		"author_id":     strconv.FormatInt(record.AuthorID, 10),
		"progress_ms":   strconv.FormatInt(record.ProgressMs, 10),
		"duration_ms":   strconv.FormatInt(record.DurationMs, 10),
		"finished":      strconv.FormatInt(record.Finished, 10),
		"source":        strconv.FormatInt(record.Source, 10),
		"device":        record.Device,
		"meta":          record.Meta,
		"first_seen_at": strconv.FormatInt(record.FirstSeenAt.Time.Unix(), 10),
		"last_seen_at":  strconv.FormatInt(record.LastSeenAt.Time.Unix(), 10),
		"sort_id":       strconv.FormatInt(sortID, 10),
	}

	itemKey := m.itemKey(record.UserID, record.MediaType, record.MediaID)
	allListKey := m.userListKey(record.UserID)
	typeListKey := m.userTypeListKey(record.UserID, record.MediaType)
	dirtyItemsKey := m.dirtyItemsKey(record.UserID)
	deleteKey := m.deleteMarkersKey(record.UserID)
	identity := historyIdentity(record.MediaType, record.MediaID)
	nowScore := record.LastSeenAt.Time.Unix()
	userIDStr := strconv.FormatInt(record.UserID, 10)

	if oldMember != "" {
		_, _ = m.rds.ZremCtx(ctx, allListKey, oldMember)
		_, _ = m.rds.ZremCtx(ctx, typeListKey, oldMember)
	}
	if err := m.rds.HmsetCtx(ctx, itemKey, fields); err != nil {
		return err
	}
	_ = m.rds.ExpireCtx(ctx, itemKey, m.cfg.DetailTTLSeconds)
	_, _ = m.rds.ZaddCtx(ctx, allListKey, nowScore, member)
	_, _ = m.rds.ZaddCtx(ctx, typeListKey, nowScore, member)
	_ = m.rds.ExpireCtx(ctx, allListKey, m.cfg.ListTTLSeconds)
	_ = m.rds.ExpireCtx(ctx, typeListKey, m.cfg.ListTTLSeconds)
	_, _ = m.rds.ZaddCtx(ctx, dirtyUsersKey, nowScore, userIDStr)
	_, _ = m.rds.SaddCtx(ctx, dirtyItemsKey, identity)
	_ = m.rds.ExpireCtx(ctx, dirtyItemsKey, m.cfg.DirtyTTLSeconds)
	_, _ = m.rds.SremCtx(ctx, deleteKey, deleteMarkerAll, deleteMarkerTypePref+strconv.FormatInt(record.MediaType, 10), deleteItemMarker(record.MediaType, record.MediaID))
	_ = m.rds.ExpireCtx(ctx, deleteKey, m.cfg.DeleteMarkerTTL)
	return nil
}

func (m *Manager) markDeleteItem(ctx context.Context, userID, mediaType, mediaID int64) error {
	userIDStr := strconv.FormatInt(userID, 10)
	identity := historyIdentity(mediaType, mediaID)
	_, err := m.rds.SaddCtx(ctx, m.deleteMarkersKey(userID), deleteItemMarker(mediaType, mediaID))
	if err != nil {
		return err
	}
	_, _ = m.rds.SaddCtx(ctx, m.dirtyItemsKey(userID), identity)
	_, _ = m.rds.ZaddCtx(ctx, dirtyUsersKey, time.Now().Unix(), userIDStr)
	_ = m.rds.ExpireCtx(ctx, m.deleteMarkersKey(userID), m.cfg.DeleteMarkerTTL)
	_ = m.rds.ExpireCtx(ctx, m.dirtyItemsKey(userID), m.cfg.DirtyTTLSeconds)
	return nil
}

func (m *Manager) markDeleteType(ctx context.Context, userID, mediaType int64) error {
	userIDStr := strconv.FormatInt(userID, 10)
	_, err := m.rds.SaddCtx(ctx, m.deleteMarkersKey(userID), deleteMarkerTypePref+strconv.FormatInt(mediaType, 10))
	if err != nil {
		return err
	}
	_, _ = m.rds.ZaddCtx(ctx, dirtyUsersKey, time.Now().Unix(), userIDStr)
	_ = m.rds.ExpireCtx(ctx, m.deleteMarkersKey(userID), m.cfg.DeleteMarkerTTL)
	return nil
}

func (m *Manager) markDeleteAll(ctx context.Context, userID int64) error {
	userIDStr := strconv.FormatInt(userID, 10)
	_, err := m.rds.SaddCtx(ctx, m.deleteMarkersKey(userID), deleteMarkerAll)
	if err != nil {
		return err
	}
	_, _ = m.rds.ZaddCtx(ctx, dirtyUsersKey, time.Now().Unix(), userIDStr)
	_ = m.rds.ExpireCtx(ctx, m.deleteMarkersKey(userID), m.cfg.DeleteMarkerTTL)
	return nil
}

func (m *Manager) membersForListKey(ctx context.Context, key string) ([]string, error) {
	return m.rds.ZrangeCtx(ctx, key, 0, -1)
}

func (m *Manager) hasDeleteMarkers(ctx context.Context, userID int64) (bool, error) {
	return m.rds.ExistsCtx(ctx, m.deleteMarkersKey(userID))
}

func (m *Manager) cacheEnabled() bool {
	return m.rds != nil
}

func (m *Manager) writeBackEnabled() bool {
	return m.cacheEnabled() && m.cfg.WriteBackEnabled
}

func (m *Manager) itemKey(userID, mediaType, mediaID int64) string {
	return fmt.Sprintf("history:item:%d:%d:%d", userID, mediaType, mediaID)
}

func (m *Manager) userListKey(userID int64) string {
	return fmt.Sprintf("history:list:%d", userID)
}

func (m *Manager) userTypeListKey(userID, mediaType int64) string {
	return fmt.Sprintf("history:list:%d:%d", userID, mediaType)
}

func (m *Manager) userSeqKey(userID int64) string {
	return fmt.Sprintf("history:seq:%d", userID)
}

func (m *Manager) dirtyItemsKey(userID int64) string {
	return fmt.Sprintf("history:dirty:items:%d", userID)
}

func (m *Manager) deleteMarkersKey(userID int64) string {
	return fmt.Sprintf("history:deleted:%d", userID)
}

func (m *Manager) flushLockKey(userID int64) string {
	return fmt.Sprintf("history:flush:lock:%d", userID)
}

func (c Config) withDefaults() Config {
	if c.ListTTLSeconds <= 0 {
		c.ListTTLSeconds = 7 * 24 * 3600
	}
	if c.DetailTTLSeconds <= 0 {
		c.DetailTTLSeconds = 7 * 24 * 3600
	}
	if c.FlushIntervalSeconds <= 0 {
		c.FlushIntervalSeconds = 60
	}
	if c.FlushBatchUsers <= 0 {
		c.FlushBatchUsers = 100
	}
	if c.FlushBatchItems <= 0 {
		c.FlushBatchItems = 200
	}
	if c.DirtyTTLSeconds <= 0 {
		c.DirtyTTLSeconds = 3600
	}
	if c.DeleteMarkerTTL <= 0 {
		c.DeleteMarkerTTL = 3600
	}
	if !c.ReadFallbackToDB {
		c.ReadFallbackToDB = true
	}
	return c
}

func historyMember(sortID, mediaType, mediaID int64) string {
	return fmt.Sprintf("%d:%d:%d", sortID, mediaType, mediaID)
}

func historyIdentity(mediaType, mediaID int64) string {
	return fmt.Sprintf("%d:%d", mediaType, mediaID)
}

func deleteItemMarker(mediaType, mediaID int64) string {
	return deleteMarkerItemPref + historyIdentity(mediaType, mediaID)
}

func parseHistoryMember(member string) (int64, int64, int64, error) {
	parts := strings.Split(member, memberSeparator)
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid history member %q", member)
	}
	sortID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	mediaType, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	mediaID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	return sortID, mediaType, mediaID, nil
}

func parseIdentity(identity string) (int64, int64, error) {
	parts := strings.Split(identity, memberSeparator)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid history identity %q", identity)
	}
	mediaType, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	mediaID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return mediaType, mediaID, nil
}

func historyRecordFromFields(fields map[string]string) (*model.HistoryRecord, error) {
	id, err := strconv.ParseInt(valueOrZero(fields["id"]), 10, 64)
	if err != nil {
		return nil, err
	}
	userID, err := strconv.ParseInt(fields["user_id"], 10, 64)
	if err != nil {
		return nil, err
	}
	mediaType, err := strconv.ParseInt(fields["media_type"], 10, 64)
	if err != nil {
		return nil, err
	}
	mediaID, err := strconv.ParseInt(fields["media_id"], 10, 64)
	if err != nil {
		return nil, err
	}
	authorID, err := strconv.ParseInt(valueOrZero(fields["author_id"]), 10, 64)
	if err != nil {
		return nil, err
	}
	progressMs, err := strconv.ParseInt(valueOrZero(fields["progress_ms"]), 10, 64)
	if err != nil {
		return nil, err
	}
	durationMs, err := strconv.ParseInt(valueOrZero(fields["duration_ms"]), 10, 64)
	if err != nil {
		return nil, err
	}
	finished, err := strconv.ParseInt(valueOrZero(fields["finished"]), 10, 64)
	if err != nil {
		return nil, err
	}
	source, err := strconv.ParseInt(valueOrZero(fields["source"]), 10, 64)
	if err != nil {
		return nil, err
	}
	firstSeenAt, err := strconv.ParseInt(valueOrZero(fields["first_seen_at"]), 10, 64)
	if err != nil {
		return nil, err
	}
	lastSeenAt, err := strconv.ParseInt(valueOrZero(fields["last_seen_at"]), 10, 64)
	if err != nil {
		return nil, err
	}

	return &model.HistoryRecord{
		ID:         id,
		UserID:     userID,
		MediaType:  mediaType,
		MediaID:    mediaID,
		Title:      fields["title"],
		Cover:      fields["cover"],
		AuthorID:   authorID,
		ProgressMs: progressMs,
		DurationMs: durationMs,
		Finished:   finished,
		Source:     source,
		Device:     fields["device"],
		Meta:       fields["meta"],
		FirstSeenAt: sql.NullTime{
			Time:  time.Unix(firstSeenAt, 0),
			Valid: firstSeenAt > 0,
		},
		LastSeenAt: sql.NullTime{
			Time:  time.Unix(lastSeenAt, 0),
			Valid: lastSeenAt > 0,
		},
	}, nil
}

func cloneHistoryRecord(in *model.HistoryRecord) *model.HistoryRecord {
	if in == nil {
		return nil
	}
	cloned := *in
	return &cloned
}

func valueOrZero(v string) string {
	if v == "" {
		return "0"
	}
	return v
}

const dirtyUsersKey = "history:dirty:users"
