package outline

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// nextPage validates the complete page before it is delivered to the sync loop.
// Outline may advertise offset+limit even for empty and short terminal pages.
// nextPath supplies pagination numbers only; it is never used as a request URL.
func (c *client) nextPage(method string, result *envelope, offset, limit, count int) (next, pageLimit int, terminal bool, err error) {
	invalid := func() (int, int, bool, error) { return 0, 0, false, errors.New("outline_scan_incomplete") }
	p := result.Pagination
	if p.Offset != nil && *p.Offset != offset {
		return invalid()
	}
	if p.Limit != nil {
		if *p.Limit <= 0 || *p.Limit > limit {
			return invalid()
		}
		limit = *p.Limit
	}
	if count > limit {
		return invalid()
	}
	next = offset + count
	if p.Total != nil {
		if *p.Total < next || (count == 0 && next < *p.Total) {
			return invalid()
		}
		terminal = next == *p.Total
	} else {
		terminal = count < limit
	}
	if p.NextPath != nil && *p.NextPath != "" {
		raw := *p.NextPath
		u, parseErr := url.Parse(raw)
		if parseErr != nil || u.User != nil || u.Opaque != "" || strings.Contains(raw, "#") ||
			u.Path != "/api/"+method || (u.Host == "" && u.Scheme != "") ||
			(u.Host != "" && u.Scheme+"://"+u.Host != c.base) {
			return invalid()
		}
		query, parseErr := url.ParseQuery(u.RawQuery)
		if parseErr != nil || len(query["offset"]) != 1 {
			return invalid()
		}
		advertised, parseErr := strconv.Atoi(query.Get("offset"))
		if parseErr != nil || advertised <= offset ||
			(advertised != next && !(terminal && advertised == offset+limit)) {
			return invalid()
		}
		if values, ok := query["limit"]; ok {
			n, parseErr := strconv.Atoi(query.Get("limit"))
			if parseErr != nil || len(values) != 1 || n != limit {
				return invalid()
			}
		}
	}
	if !terminal && next <= offset {
		return invalid()
	}
	return next, limit, terminal, nil
}
