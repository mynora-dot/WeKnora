package outline

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tencent/WeKnora/internal/datasource"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Tencent/WeKnora/internal/types"
	"github.com/stretchr/testify/require"
)

// This models the reported 9/25 response without retaining instance IDs or secrets.
func TestCollectionDiscoveryUsesTerminalPagination(t *testing.T) {
	f, cfg := testFixture(t)
	f.terminalNextPath = true
	for i := 0; i < 9; i++ {
		f.collections = append(f.collections, collection{ID: fmt.Sprintf("%08d-1111-4111-8111-111111111111", i), Name: fmt.Sprintf("Collection %d", i)})
	}
	cfg.ResourceIDs, cfg.SyncDeletions = nil, false
	c := NewConnector()
	require.NoError(t, c.Validate(context.Background(), cfg))
	resources, err := c.ListResources(context.Background(), cfg, "")
	require.NoError(t, err)
	require.Len(t, resources, 9)
}

func TestPaginationMetadataValidation(t *testing.T) {
	c := &client{base: "https://outline.example.com"}
	for _, tc := range []struct {
		name, pagination string
		count            int
	}{
		{"wrong offset", `{"offset":1,"total":1}`, 1},
		{"invalid limit", `{"limit":0,"total":1}`, 1},
		{"oversized page", `{"limit":1,"total":2}`, 2},
		{"duplicate offset", `{"nextPath":"/api/documents.list?offset=25&offset=50"}`, 1},
		{"duplicate limit", `{"nextPath":"/api/documents.list?offset=25&limit=25&limit=1"}`, 1},
		{"userinfo", `{"nextPath":"https://user:pass@outline.example.com/api/documents.list?offset=25"}`, 1},
		{"fragment", `{"nextPath":"/api/documents.list?offset=25#x"}`, 1},
		{"wrong endpoint", `{"nextPath":"/api/collections.list?offset=25"}`, 1},
		{"malformed query", `{"nextPath":"/api/documents.list?offset=25&limit=%ZZ"}`, 1},
		{"negative total", `{"total":-1}`, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var e envelope
			require.NoError(t, json.Unmarshal([]byte(`{"pagination":`+tc.pagination+`}`), &e))
			_, _, _, err := c.nextPage("documents.list", &e, 0, 25, tc.count)
			require.ErrorContains(t, err, "outline_scan_incomplete")
		})
	}
	for _, path := range []string{"", "/api/documents.list?offset=1&limit=1"} {
		var e envelope
		b, err := json.Marshal(map[string]interface{}{"pagination": map[string]interface{}{"limit": 1, "offset": 0, "total": 2, "nextPath": path}})
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(b, &e))
		next, limit, terminal, err := c.nextPage("documents.list", &e, 0, 25, 1)
		require.NoError(t, err)
		require.Equal(t, 1, next)
		require.Equal(t, 1, limit)
		require.False(t, terminal)
	}
}

func TestDeletionRetainsDocumentWhenDetailStateIsMissing(t *testing.T) {
	f, cfg := testFixture(t)
	f.docs["a"] = sample("a", collectionA)
	f.listed[collectionA] = []string{"a"}
	c := NewConnector()
	previous, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
	require.NoError(t, err)
	f.listed[collectionA] = nil
	f.deleted = []string{"a"}
	f.detailOverride["a"] = json.RawMessage(`{"id":"a","text":"still readable"}`)
	h := &acknowledger{}
	next, err := c.FetchStream(context.Background(), cfg, previous, h)
	require.Error(t, err)
	require.Empty(t, h.items)
	state, err := parseCursor(next)
	require.NoError(t, err)
	require.Contains(t, state.Documents, "a")
}

func TestMissingListStateRequiresDetailBeforeAcceptance(t *testing.T) {
	f, cfg := testFixture(t)
	f.docs["a"] = sample("a", collectionA)
	f.listed[collectionA] = []string{"a"}
	f.listOverride["a"] = json.RawMessage(`{"id":"a","text":"summary"}`)
	h := &acknowledger{}
	next, err := NewConnector().FetchStream(context.Background(), cfg, nil, h)
	require.NoError(t, err)
	require.Equal(t, 1, f.detailCalls)
	require.Len(t, h.items, 1)
	require.Contains(t, string(h.items[0].Content), "hello")
	state, err := parseCursor(next)
	require.NoError(t, err)
	require.Contains(t, state.Documents, "a")
}

func TestSourceTimestampsRemainInKnowledgeMetadata(t *testing.T) {
	d := sample("a", collectionA)
	d.CreatedAt = "2026-08-01T08:00:00+08:00"
	item, _, err := mappedItem(d, identity{BaseURL: "https://outline.example.com"})
	require.NoError(t, err)
	require.Equal(t, "2026-08-01T00:00:00Z", item.Metadata["source_created_at"])
	require.Equal(t, "2026-09-01T00:00:00Z", item.Metadata["source_updated_at"])
	require.Equal(t, types.ChannelOutline, item.Metadata["channel"])
}

func TestOversizedListReducesLimitAtSameOffset(t *testing.T) {
	_, cfg := testFixture(t)
	c, err := newClient(cfg)
	require.NoError(t, err)
	oversized := strings.Repeat("x", maxResponse+1)
	var offsets, limits []int
	c.http.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		var request struct{ Offset, Limit int }
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		offsets = append(offsets, request.Offset)
		limits = append(limits, request.Limit)
		data := oversized
		if request.Limit <= 6 {
			rows := []map[string]string{}
			for i := request.Offset; i < min(9, request.Offset+request.Limit); i++ {
				rows = append(rows, map[string]string{"id": fmt.Sprint(i)})
			}
			raw, err := json.Marshal(map[string]interface{}{"data": rows, "pagination": map[string]int{"total": 9}})
			require.NoError(t, err)
			data = string(raw)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(data))}, nil
	})
	count := 0
	require.NoError(t, c.walk(context.Background(), "documents.list", map[string]interface{}{}, func(rows []json.RawMessage) error { count += len(rows); return nil }))
	require.Equal(t, []int{0, 0, 0, 6}, offsets)
	require.Equal(t, []int{25, 12, 6, 6}, limits)
	require.Equal(t, 9, count)
}

func TestPendingFailuresAreNotTruncatedWithErrorSamples(t *testing.T) {
	f, cfg := testFixture(t)
	for i := 0; i < 150; i++ {
		id := fmt.Sprintf("doc-%d", i)
		d := sample(id, collectionA)
		d.Text = nil
		f.docs[id] = d
		f.listed[collectionA] = append(f.listed[collectionA], id)
	}
	c := NewConnector()
	next, err := c.FetchStream(context.Background(), cfg, nil, &acknowledger{})
	var partial *datasource.PartialFetchError
	require.ErrorAs(t, err, &partial)
	require.Len(t, partial.Details, 100)
	state, err := parseCursor(next)
	require.NoError(t, err)
	require.Len(t, state.PendingUpserts, 150)
	require.Empty(t, state.Documents)
	for id := range f.docs {
		f.docs[id] = sample(id, collectionA)
	}
	h := &acknowledger{}
	next, err = c.FetchStream(context.Background(), cfg, next, h)
	require.NoError(t, err)
	state, err = parseCursor(next)
	require.NoError(t, err)
	require.Len(t, state.Documents, 150)
	require.Empty(t, state.PendingUpserts)
	require.Len(t, h.items, 150)
}
