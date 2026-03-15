package service

import (
	"context"
	"fmt"

	"rasanta/pkg/gmapsparser"
)

type placeService struct {
	repository Repository
}

func NewPlaceService(repository Repository) PlaceService {
	return &placeService{repository: repository}
}

func (s *placeService) ParseRawText(rawText, hl string) ([]map[string]any, error) {
	return gmapsparser.ParsePlaces(rawText, hl)
}

func (s *placeService) FetchSearch(ctx context.Context, params SearchParams) (int, []map[string]any, error) {
	result, err := s.repository.FetchSearch(ctx, params)
	if err != nil {
		// Tetap bawa status code upstream kalau tersedia, biar observability lebih jelas.
		if result != nil {
			return result.StatusCode, nil, err
		}
		return 0, nil, err
	}

	places, err := gmapsparser.ParsePlaces(result.Body, params.HL)
	if err != nil {
		return result.StatusCode, nil, err
	}
	if len(places) == 0 {
		return result.StatusCode, nil, fmt.Errorf("place block tidak ditemukan")
	}

	return result.StatusCode, places, nil
}

func (s *placeService) FetchDetail(ctx context.Context, params DetailParams) (int, map[string]any, error) {
	result, err := s.repository.FetchDetail(ctx, params)
	if err != nil {
		if result != nil {
			return result.StatusCode, nil, err
		}
		return 0, nil, err
	}

	place, err := gmapsparser.ParseFirstPlace(result.Body, params.HL)
	if err != nil {
		// Parsing error dianggap error bisnis karena payload upstream tidak sesuai ekspektasi parser.
		return result.StatusCode, nil, err
	}

	return result.StatusCode, place, nil
}
