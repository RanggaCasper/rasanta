package gmapsparser

import (
	"math"
	"net/url"
	"regexp"
	"strings"
)

var (
	dayMap = map[string]string{
		"monday":    "monday",
		"tuesday":   "tuesday",
		"wednesday": "wednesday",
		"thursday":  "thursday",
		"friday":    "friday",
		"saturday":  "saturday",
		"sunday":    "sunday",
		"senin":     "monday",
		"selasa":    "tuesday",
		"rabu":      "wednesday",
		"kamis":     "thursday",
		"jumat":     "friday",
		"jum'at":    "friday",
		"sabtu":     "saturday",
		"minggu":    "sunday",
	}

	dataIDPattern  = regexp.MustCompile(`(?i)^0x[0-9a-f]+:0x[0-9a-f]+$`)
	phonePattern   = regexp.MustCompile(`^[+()\-\d\s]{7,}$`)
)

func walk(node any, visit func(any) bool) bool {
	if !visit(node) {
		return false
	}

	switch typed := node.(type) {
	case []any:
		for _, item := range typed {
			if !walk(item, visit) {
				return false
			}
		}
	case map[string]any:
		for _, item := range typed {
			if !walk(item, visit) {
				return false
			}
		}
	}

	return true
}

func safeGet(seq any, indexes ...int) any {
	current := seq
	for _, idx := range indexes {
		items, ok := current.([]any)
		if !ok || idx < 0 || idx >= len(items) {
			return nil
		}
		current = items[idx]
	}
	return current
}

func asSlice(value any) ([]any, bool) {
	items, ok := value.([]any)
	return items, ok
}

func asString(value any) (string, bool) {
	text, ok := value.(string)
	return text, ok
}

func asInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int8:
		return int(typed), true
	case int16:
		return int(typed), true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float32:
		if math.Trunc(float64(typed)) == float64(typed) {
			return int(typed), true
		}
	case float64:
		if math.Trunc(typed) == typed {
			return int(typed), true
		}
	}
	return 0, false
}

func asFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	}
	return 0, false
}

func toStringSlice(value any) []string {
	items, ok := asSlice(value)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := asString(item); ok {
			result = append(result, text)
		}
	}
	return result
}

func normalizeText(value any) *string {
	text, ok := asString(value)
	if !ok {
		return nil
	}

	cleaned := strings.Join(strings.Fields(text), " ")
	cleaned = strings.TrimSpace(cleaned)
	if cleaned == "" {
		return nil
	}

	if len(cleaned) >= 2 && strings.HasPrefix(cleaned, `"`) && strings.HasSuffix(cleaned, `"`) {
		cleaned = strings.TrimSpace(cleaned[1 : len(cleaned)-1])
	}

	if cleaned == "" {
		return nil
	}

	return &cleaned
}

func stringOrNil(value any) *string {
	text, ok := asString(value)
	if !ok {
		return nil
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	return &text
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func isValidCoordinatePair(latValue, lngValue any) bool {
	lat, okLat := asFloat(latValue)
	lng, okLng := asFloat(lngValue)
	if !okLat || !okLng {
		return false
	}
	if lat < -90 || lat > 90 {
		return false
	}
	if lng < -180 || lng > 180 {
		return false
	}
	return true
}

func isGoogleHost(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	return strings.Contains(host, "google.") ||
		strings.Contains(host, "gstatic.") ||
		strings.Contains(host, "googleusercontent.com") ||
		strings.Contains(host, "googleapis.com")
}

func findFirstURL(node any, predicate func(string) bool) *string {
	var found *string

	walk(node, func(item any) bool {
		if found != nil {
			return false
		}

		text, ok := asString(item)
		if !ok {
			return true
		}

		if !strings.HasPrefix(text, "http://") && !strings.HasPrefix(text, "https://") {
			return true
		}

		if predicate(text) {
			copy := text
			found = &copy
			return false
		}

		return true
	})

	return found
}
