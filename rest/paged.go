package rest

import (
	"strconv"
)

// Filter to get paged results
type Filter struct {
	Limit    int64  `json:"limit"`
	Offset   int64  `json:"offset"`
	FromID   string `json:"fromId"`
	ToID     string `json:"toId"`
	FromDate int64  `json:"fromDate"`
	ToDate   int64  `json:"toDate"`
}

// Paging params
type Paging struct {
	Limit   int64  `json:"limit"`
	Offset  int64  `json:"offset"`
	Total   int    `json:"total"`
	FirstID string `json:"firstId"`
	LastID  string `json:"lastId"`
}

// Page result
type Page struct {
	Paging *Paging     `json:"paging"`
	Result interface{} `json:"result"`
}

// PagedEndpointMiddleware adds filters to get a paged result
func PagedEndpointMiddleware(handler HandlerFunc) HandlerFunc {
	return func(r *Request) *Response {
		limitStr := r.QueryParam("limit")
		if limitStr != "" {
			limit, err := strconv.ParseInt(limitStr, 10, 64)
			if err != nil {
				return BadRequest(ErrInvalidNumberParam("limit"))
			}
			r.Filter.Limit = limit
		}

		offsetStr := r.QueryParam("offset")
		if offsetStr != "" {
			offset, err := strconv.ParseInt(offsetStr, 10, 64)
			if err != nil {
				return BadRequest(ErrInvalidNumberParam("offset"))
			}
			r.Filter.Offset = offset
		}

		fromID := r.QueryParam("fromId")
		if fromID != "" {
			r.Filter.FromID = fromID
		}

		toID := r.QueryParam("toId")
		if len(toID) > 0 {
			r.Filter.ToID = toID
		}

		strFromDate := r.QueryParam("fromDate")
		if strFromDate != "" {
			fromDate, err := strconv.ParseInt(strFromDate, 10, 64)
			if err != nil {
				return BadRequest(ErrInvalidNumberParam("fromDate"))
			}
			r.Filter.FromDate = fromDate
		}

		strToDate := r.QueryParam("toDate")
		if strToDate != "" {
			toDate, err := strconv.ParseInt(strToDate, 10, 64)
			if err != nil {
				return BadRequest(ErrInvalidNumberParam("toDate"))
			}
			r.Filter.ToDate = toDate
		}

		return handler(r)
	}
}
