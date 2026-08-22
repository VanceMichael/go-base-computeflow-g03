package httpapi

import (
	"github.com/VanceMichael/harborflow/internal/domain"
	"net/http"
	"strconv"
)

func pageFromRequest(r *http.Request) (domain.Page, error) {
	limit := 50
	offset := 0
	var err error
	if v := r.URL.Query().Get("limit"); v != "" {
		limit, err = strconv.Atoi(v)
		if err != nil {
			return domain.Page{}, domain.ErrInvalid
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		offset, err = strconv.Atoi(v)
		if err != nil {
			return domain.Page{}, domain.ErrInvalid
		}
	}
	return domain.NewPage(limit, offset)
}
func pageHeaders(w http.ResponseWriter, info domain.PageInfo) {
	w.Header().Set("X-Total-Count", strconv.Itoa(info.Total))
	w.Header().Set("X-Page-Limit", strconv.Itoa(info.Limit))
	w.Header().Set("X-Page-Offset", strconv.Itoa(info.Offset))
}
