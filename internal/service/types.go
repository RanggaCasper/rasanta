package service

import "context"

type FetchResult struct {
	StatusCode int
	Body       string
}

type SearchParams struct {
	Query    string
	Lat      float64
	Lng      float64
	HL       string
	GL       string
	AuthUser string
	Limit    int
	UseSAW   bool

	RatingWeight  float64
	ReviewsWeight float64
	PriceWeight   float64
}

type DetailParams struct {
	PB       string
	Q        string
	HL       string
	GL       string
	AuthUser string
}

type Repository interface {
	FetchSearch(ctx context.Context, params SearchParams) (*FetchResult, error)
	FetchDetail(ctx context.Context, params DetailParams) (*FetchResult, error)
}

type PlaceService interface {
	FetchSearch(ctx context.Context, params SearchParams) (int, []map[string]any, error)
	FetchDetail(ctx context.Context, params DetailParams) (int, map[string]any, error)
}
