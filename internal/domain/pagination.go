package domain

import "fmt"

type Page struct {
	Limit  int
	Offset int
}

func NewPage(limit, offset int) (Page, error) {
	if limit <= 0 || limit > 200 {
		return Page{}, fmt.Errorf("%w: limit", ErrInvalid)
	}
	if offset < 0 {
		return Page{}, fmt.Errorf("%w: offset", ErrInvalid)
	}
	return Page{Limit: limit, Offset: offset}, nil
}

type PageInfo struct {
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

func NewPageInfo(total int, p Page) PageInfo {
	return PageInfo{Total: total, Limit: p.Limit, Offset: p.Offset, HasMore: p.Offset+p.Limit < total}
}
