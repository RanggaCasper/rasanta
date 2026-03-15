package service

import (
	"sort"
	"strconv"
	"strings"
)

type sawRow struct {
	place       map[string]any
	rating      float64
	reviews     float64
	price       float64
	ratingNorm  float64
	reviewsNorm float64
	priceNorm   float64
	score       float64
}

func rankPlacesWithSAW(places []map[string]any, ratingWeight, reviewsWeight, priceWeight float64) []map[string]any {
	if len(places) == 0 {
		return places
	}

	ratingWeight, reviewsWeight, priceWeight = normalizeWeights(ratingWeight, reviewsWeight, priceWeight)

	rows := make([]sawRow, 0, len(places))
	maxRating := 0.0
	maxReviews := 0.0
	minPrice := 0.0

	for _, place := range places {
		rating := toFloatValue(place["rating"])
		reviews := toFloatValue(place["reviews"])
		price := toPriceValue(place["price"])

		if rating > maxRating {
			maxRating = rating
		}
		if reviews > maxReviews {
			maxReviews = reviews
		}
		if price > 0 && (minPrice == 0 || price < minPrice) {
			minPrice = price
		}

		rows = append(rows, sawRow{
			place:   place,
			rating:  rating,
			reviews: reviews,
			price:   price,
		})
	}

	for i := range rows {
		if maxRating > 0 && rows[i].rating > 0 {
			rows[i].ratingNorm = rows[i].rating / maxRating
		}
		if maxReviews > 0 && rows[i].reviews > 0 {
			rows[i].reviewsNorm = rows[i].reviews / maxReviews
		}
		if minPrice > 0 && rows[i].price > 0 {
			rows[i].priceNorm = minPrice / rows[i].price
		}

		rows[i].score = (ratingWeight * rows[i].ratingNorm) +
			(reviewsWeight * rows[i].reviewsNorm) +
			(priceWeight * rows[i].priceNorm)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].score > rows[j].score
	})

	ranked := make([]map[string]any, 0, len(rows))
	for idx := range rows {
		rows[idx].place["saw_score"] = rows[idx].score
		rows[idx].place["saw_rank"] = idx + 1
		rows[idx].place["saw_normalized"] = map[string]any{
			"rating":  rows[idx].ratingNorm,
			"reviews": rows[idx].reviewsNorm,
			"price":   rows[idx].priceNorm,
		}
		ranked = append(ranked, rows[idx].place)
	}

	return ranked
}

func normalizeWeights(ratingWeight, reviewsWeight, priceWeight float64) (float64, float64, float64) {
	if ratingWeight < 0 {
		ratingWeight = 0
	}
	if reviewsWeight < 0 {
		reviewsWeight = 0
	}
	if priceWeight < 0 {
		priceWeight = 0
	}

	total := ratingWeight + reviewsWeight + priceWeight
	if total == 0 {
		return 0.6, 0.3, 0.1
	}

	return ratingWeight / total, reviewsWeight / total, priceWeight / total
}

func toFloatValue(v any) float64 {
	switch value := v.(type) {
	case int:
		return float64(value)
	case int32:
		return float64(value)
	case int64:
		return float64(value)
	case float32:
		return float64(value)
	case float64:
		return value
	case string:
		clean := strings.TrimSpace(strings.ReplaceAll(value, ",", ""))
		if clean == "" {
			return 0
		}
		parsed, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		return 0
	}
}

func toPriceValue(v any) float64 {
	switch value := v.(type) {
	case int, int32, int64, float32, float64:
		return toFloatValue(value)
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return 0
		}
		if strings.Contains(trimmed, "$") {
			count := strings.Count(trimmed, "$")
			if count > 0 {
				return float64(count)
			}
		}
		return toFloatValue(trimmed)
	default:
		return 0
	}
}
