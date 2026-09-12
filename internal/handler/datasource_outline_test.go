package handler

import (
	"context"
	"github.com/Tencent/WeKnora/internal/datasource"
	"github.com/Tencent/WeKnora/internal/types"
	"github.com/Tencent/WeKnora/internal/types/interfaces"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeForceFull(t *testing.T) {
	for _, tc := range []struct {
		body          string
		full, invalid bool
	}{
		{"", false, false}, {"{}", false, false}, {`{"force_full":true}`, true, false},
		{`{"force_full":false}`, false, false}, {`{"force_full":"true"}`, false, true},
		{`{"force_full":1}`, false, true}, {`{"force_full":null}`, false, true},
		{"null", false, true}, {"[]", false, true}, {"{}{}", false, true},
	} {
		t.Run(tc.body, func(t *testing.T) {
			full, err := decodeForceFull(strings.NewReader(tc.body))
			if (err != nil) != tc.invalid || full != tc.full {
				t.Fatalf("full=%v err=%v", full, err)
			}
		})
	}
}

type outlineRequestService struct {
	interfaces.DataSourceService
	created *types.DataSource
	full    bool
	syncErr error
}

func (s *outlineRequestService) CreateDataSource(_ context.Context, ds *types.DataSource) (*types.DataSource, error) {
	s.created = ds
	return ds, nil
}
func (s *outlineRequestService) GetDataSource(context.Context, string) (*types.DataSource, error) {
	return &types.DataSource{ID: "ds", TenantID: 1, KnowledgeBaseID: "kb"}, nil
}
func (s *outlineRequestService) ManualSyncWithOptions(_ context.Context, _ string, full bool) (*types.SyncLog, error) {
	s.full = full
	return &types.SyncLog{ID: "log"}, s.syncErr
}

func TestOutlineCreateDefaultsAndExplicitDisabledOptions(t *testing.T) {
	for _, tc := range []struct {
		name, extra string
		deletion    bool
		schedule    string
	}{
		{"defaults", "", true, "0 0 */6 * * *"},
		{"disabled", `,"sync_deletions":false,"sync_schedule":""`, false, ""},
		{"draft", `,"status":"paused","sync_deletions":false`, false, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc := &outlineRequestService{}
			kb := &stubKBServiceForDS{getByID: func(context.Context, string) (*types.KnowledgeBase, error) {
				return &types.KnowledgeBase{ID: "kb", TenantID: 1}, nil
			}}
			h := NewDataSourceHandler(svc, kb)
			r := newDataSourceTestRouter(h)
			r.POST("/datasource", h.CreateDataSource)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/datasource", strings.NewReader(`{"type":"outline","knowledge_base_id":"kb"`+tc.extra+`}`))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, withDSCtx(req, 1))
			require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
			require.Equal(t, tc.deletion, svc.created.SyncDeletions)
			require.Equal(t, tc.schedule, svc.created.SyncSchedule)
			require.Equal(t, types.SyncModeIncremental, svc.created.SyncMode)
			require.Equal(t, "overwrite", svc.created.ConflictStrategy)
		})
	}
}

func TestOutlineManualSyncOptionsAndConflict(t *testing.T) {
	for _, conflict := range []bool{false, true} {
		svc := &outlineRequestService{}
		if conflict {
			svc.syncErr = datasource.ErrSyncRunning
		}
		kb := &stubKBServiceForDS{getByID: func(context.Context, string) (*types.KnowledgeBase, error) {
			return &types.KnowledgeBase{ID: "kb", TenantID: 1}, nil
		}}
		h := NewDataSourceHandler(svc, kb)
		r := newDataSourceTestRouter(h)
		r.POST("/datasource/:id/sync", h.ManualSync)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/datasource/ds/sync", strings.NewReader(`{"force_full":true}`))
		r.ServeHTTP(w, withDSCtx(req, 1))
		require.True(t, svc.full)
		if conflict {
			require.Equal(t, http.StatusConflict, w.Code)
			require.Contains(t, w.Body.String(), "datasource_sync_running")
		} else {
			require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		}
	}
}
