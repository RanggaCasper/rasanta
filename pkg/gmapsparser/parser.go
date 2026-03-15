package gmapsparser

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParsePlaces(text, hl string) ([]map[string]any, error) {
	payload, err := parsePayloadFromText(text)
	if err != nil {
		return nil, err
	}

	placeBlocks := findPlaceBlocks(payload)
	result := make([]map[string]any, 0, len(placeBlocks))
	for index, block := range placeBlocks {
		result = append(result, toPlaceOutput(block, index+1, defaultString(hl, "en")))
	}

	return result, nil
}

func ParseFirstPlace(text, hl string) (map[string]any, error) {
	places, err := ParsePlaces(text, hl)
	if err != nil {
		return nil, err
	}
	if len(places) == 0 {
		return nil, fmt.Errorf("place block tidak ditemukan")
	}
	return places[0], nil
}

func cleanResponse(text string) string {
	cleaned := strings.TrimLeft(text, " \t\r\n")
	if strings.HasPrefix(cleaned, ")]}'") {
		return cleaned[4:]
	}
	return cleaned
}

func parsePayloadFromText(rawText string) (any, error) {
	base := cleanResponse(rawText)
	jsonText := strings.TrimSpace(strings.SplitN(base, "/*", 2)[0])
	if jsonText == "" {
		return nil, fmt.Errorf("response kosong atau tidak valid")
	}

	switch jsonText[0] {
	case '{':
		var envelope map[string]any
		if err := json.Unmarshal([]byte(jsonText), &envelope); err != nil {
			return nil, fmt.Errorf("decode envelope: %w", err)
		}

		detailRaw, ok := envelope["d"].(string)
		if ok {
			payloadText := strings.TrimSpace(cleanResponse(detailRaw))
			if payloadText == "" {
				return nil, fmt.Errorf("payload 'd' kosong")
			}

			var payload any
			if err := json.Unmarshal([]byte(payloadText), &payload); err != nil {
				return nil, fmt.Errorf("decode payload: %w", err)
			}
			return payload, nil
		}

		if payload, ok := envelope["d"]; ok {
			return payload, nil
		}

		return nil, fmt.Errorf("field 'd' tidak ditemukan")
	case '[':
		var payload any
		if err := json.Unmarshal([]byte(jsonText), &payload); err != nil {
			return nil, fmt.Errorf("decode payload: %w", err)
		}
		return payload, nil
	default:
		return nil, fmt.Errorf("format response tidak dikenali")
	}
}

func findPlaceBlocks(payload any) [][]any {
	blocks := make([][]any, 0)
	seenKeys := make(map[string]struct{})

	walk(payload, func(node any) bool {
		items, ok := node.([]any)
		if !ok || len(items) < 120 {
			return true
		}

		dataID := ""
		placeID := ""
		for _, value := range toStringSlice(items) {
			if dataID == "" && dataIDPattern.MatchString(value) {
				dataID = value
			}
			if placeID == "" && strings.HasPrefix(value, "ChIJ") {
				placeID = value
			}
			if dataID != "" && placeID != "" {
				break
			}
		}

		if dataID == "" || placeID == "" {
			return true
		}

		key := dataID + "|" + placeID
		if _, exists := seenKeys[key]; exists {
			return true
		}

		seenKeys[key] = struct{}{}
		blocks = append(blocks, items)
		return true
	})

	return blocks
}
