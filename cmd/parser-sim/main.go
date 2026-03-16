package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var dataIDPattern = regexp.MustCompile(`(?i)^0x[0-9a-f]+:0x[0-9a-f]+$`)

type options struct {
	inputPath  string
	outPath    string
	place      int
	pathRaw    string
	findText   string
	findLimit  int
	ignoreCase bool
	pretty     bool
	showPlace  bool
}

type traceStep struct {
	Depth           int    `json:"depth"`
	Index           int    `json:"index"`
	ContainerType   string `json:"container_type"`
	ContainerLength int    `json:"container_length"`
	Found           bool   `json:"found"`
	NextType        string `json:"next_type,omitempty"`
}

type sliceTraceResult struct {
	PathRaw   string      `json:"path_raw"`
	Path      []int       `json:"path"`
	Found     bool        `json:"found"`
	ValueType string      `json:"value_type,omitempty"`
	Value     any         `json:"value,omitempty"`
	Error     string      `json:"error,omitempty"`
	Steps     []traceStep `json:"steps"`
}

type textMatch struct {
	Path       string `json:"path"`
	PathTokens []any  `json:"path_tokens"`
	Value      string `json:"value"`
}

type textSearchResult struct {
	Query      string      `json:"query"`
	IgnoreCase bool        `json:"ignore_case"`
	Limit      int         `json:"limit"`
	Count      int         `json:"count"`
	Matches    []textMatch `json:"matches"`
}

type simResponse struct {
	OK            bool   `json:"ok"`
	Input         string `json:"input"`
	Output        string `json:"output"`
	PlaceBlocks   int    `json:"place_blocks"`
	TraceIncluded bool   `json:"trace_included"`
	ShowPlace     bool   `json:"show_place"`
	FindText      string `json:"find_text,omitempty"`
	MatchCount    int    `json:"match_count,omitempty"`
	FirstPath     string `json:"first_path,omitempty"`
}

type simErrorResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func main() {
	opts, err := parseFlags()
	if err != nil {
		fail(err, true)
	}

	rawBytes, err := os.ReadFile(opts.inputPath)
	if err != nil {
		fail(fmt.Errorf("baca input: %w", err), opts.pretty)
	}

	rawText := string(rawBytes)
	payload, err := parsePayloadFromText(rawText)
	if err != nil {
		fail(fmt.Errorf("decode payload gagal: %w", err), opts.pretty)
	}

	blocks := findPlaceBlocks(payload)

	dump := map[string]any{
		"input":        opts.inputPath,
		"payload":      payload,
		"place_blocks": len(blocks),
	}

	if opts.showPlace || strings.TrimSpace(opts.pathRaw) != "" {
		if len(blocks) == 0 {
			fail(errors.New("place block tidak ditemukan di payload mentah"), opts.pretty)
		}

		placeIndex := opts.place - 1
		if placeIndex < 0 || placeIndex >= len(blocks) {
			fail(fmt.Errorf("place index di luar range: %d (1..%d)", opts.place, len(blocks)), opts.pretty)
		}

		selectedBlock := blocks[placeIndex]
		dump["selected_place_index"] = opts.place

		if opts.showPlace {
			dump["selected_place_block"] = selectedBlock
		}

		if strings.TrimSpace(opts.pathRaw) != "" {
			path, err := parsePath(opts.pathRaw)
			if err != nil {
				fail(fmt.Errorf("path tidak valid: %w", err), opts.pretty)
			}

			value, found, steps := traceGetAny(selectedBlock, path)
			trace := &sliceTraceResult{
				PathRaw: opts.pathRaw,
				Path:    path,
				Found:   found,
				Steps:   steps,
			}

			if found {
				trace.ValueType = fmt.Sprintf("%T", value)
				trace.Value = value
			} else {
				trace.Error = "path tidak ditemukan"
			}

			dump["slice_trace"] = trace
		}
	}

	search := findTextInPayload(payload, opts.findText, opts.findLimit, opts.ignoreCase)
	if search != nil {
		dump["text_search"] = search
	}

	if err := writeJSONFile(opts.outPath, dump, opts.pretty); err != nil {
		fail(fmt.Errorf("simpan file output gagal: %w", err), opts.pretty)
	}

	firstPath := ""
	matchCount := 0
	if search != nil {
		matchCount = search.Count
		if len(search.Matches) > 0 {
			firstPath = search.Matches[0].Path
		}
	}

	writeJSON(simResponse{
		OK:            true,
		Input:         opts.inputPath,
		Output:        opts.outPath,
		PlaceBlocks:   len(blocks),
		TraceIncluded: strings.TrimSpace(opts.pathRaw) != "",
		ShowPlace:     opts.showPlace,
		FindText:      strings.TrimSpace(opts.findText),
		MatchCount:    matchCount,
		FirstPath:     firstPath,
	}, opts.pretty)
}

func parseFlags() (options, error) {
	var opts options

	flag.StringVar(&opts.inputPath, "input", "", "path file response mentah")
	flag.StringVar(&opts.outPath, "out", "", "path output JSON (default: <input>.raw.json)")
	flag.IntVar(&opts.place, "place", 1, "nomor place (1-based)")
	flag.StringVar(&opts.pathRaw, "path", "", "index path slice, contoh: 178 atau 178.0.2")
	flag.StringVar(&opts.findText, "find-text", "", "cari teks di seluruh payload dan tampilkan path yang cocok")
	flag.IntVar(&opts.findLimit, "find-limit", 20, "batas jumlah path hasil pencarian teks")
	flag.BoolVar(&opts.ignoreCase, "find-ignore-case", false, "pencarian teks case-insensitive")
	flag.BoolVar(&opts.pretty, "pretty", true, "output JSON pretty")
	flag.BoolVar(&opts.showPlace, "show-place", false, "tampilkan raw selected place block ke file output")
	flag.Parse()

	if strings.TrimSpace(opts.inputPath) == "" {
		return options{}, errors.New("flag -input wajib diisi")
	}
	if opts.place < 1 {
		return options{}, errors.New("flag -place minimal 1")
	}
	if opts.findLimit < 1 {
		return options{}, errors.New("flag -find-limit minimal 1")
	}

	if strings.TrimSpace(opts.outPath) == "" {
		opts.outPath = defaultOutputPath(opts.inputPath)
	}

	return opts, nil
}

func defaultOutputPath(inputPath string) string {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		name = "output"
	}

	return filepath.Join(dir, name+".raw.json")
}

func parsePath(raw string) ([]int, error) {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '.' || r == '/' || r == ',' || r == ' '
	})
	if len(parts) == 0 {
		return nil, errors.New("path kosong")
	}

	indexes := make([]int, 0, len(parts))
	for _, part := range parts {
		idx, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("index %q bukan angka", part)
		}
		if idx < 0 {
			return nil, fmt.Errorf("index %d tidak boleh negatif", idx)
		}
		indexes = append(indexes, idx)
	}

	return indexes, nil
}

func traceGetAny(seq any, path []int) (any, bool, []traceStep) {
	steps := make([]traceStep, 0, len(path))
	current := seq
	for depth, idx := range path {
		items, ok := current.([]any)
		step := traceStep{
			Depth:         depth,
			Index:         idx,
			ContainerType: fmt.Sprintf("%T", current),
			Found:         false,
		}
		if ok {
			step.ContainerLength = len(items)
		}

		if !ok || idx < 0 || idx >= len(items) {
			steps = append(steps, step)
			return nil, false, steps
		}

		step.Found = true
		current = items[idx]
		step.NextType = fmt.Sprintf("%T", current)
		steps = append(steps, step)
	}
	return current, true, steps
}

func safeGetAny(seq any, path []int) (any, bool) {
	value, found, _ := traceGetAny(seq, path)
	return value, found
}

func writeJSON(value any, pretty bool) {
	var (
		payload []byte
		err     error
	)

	if pretty {
		payload, err = json.MarshalIndent(value, "", "  ")
	} else {
		payload, err = json.Marshal(value)
	}
	if err != nil {
		fallback, _ := json.Marshal(simErrorResponse{OK: false, Error: "marshal output gagal: " + err.Error()})
		fmt.Println(string(fallback))
		os.Exit(1)
	}

	fmt.Println(string(payload))
}

func writeJSONFile(path string, value any, pretty bool) error {
	var (
		payload []byte
		err     error
	)

	if pretty {
		payload, err = json.MarshalIndent(value, "", "  ")
	} else {
		payload, err = json.Marshal(value)
	}
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, payload, 0o644)
}

func findTextInPayload(payload any, query string, limit int, ignoreCase bool) *textSearchResult {
	needle := strings.TrimSpace(query)
	if needle == "" {
		return nil
	}

	normalizedNeedle := needle
	if ignoreCase {
		normalizedNeedle = strings.ToLower(needle)
	}

	matches := make([]textMatch, 0)
	var walkWithPath func(node any, path []any)
	walkWithPath = func(node any, path []any) {
		if len(matches) >= limit {
			return
		}

		switch typed := node.(type) {
		case string:
			candidate := typed
			target := candidate
			if ignoreCase {
				target = strings.ToLower(candidate)
			}

			if strings.Contains(target, normalizedNeedle) {
				pathTokens := make([]any, 0, len(path)+1)
				pathTokens = append(pathTokens, "payload")
				pathTokens = append(pathTokens, append([]any{}, path...)...)

				matches = append(matches, textMatch{
					Path:       renderPath(path),
					PathTokens: pathTokens,
					Value:      candidate,
				})
			}
		case []any:
			for idx, item := range typed {
				walkWithPath(item, appendPath(path, idx))
				if len(matches) >= limit {
					return
				}
			}
		case map[string]any:
			for key, item := range typed {
				walkWithPath(item, appendPath(path, key))
				if len(matches) >= limit {
					return
				}
			}
		}
	}

	walkWithPath(payload, nil)

	return &textSearchResult{
		Query:      needle,
		IgnoreCase: ignoreCase,
		Limit:      limit,
		Count:      len(matches),
		Matches:    matches,
	}
}

func appendPath(path []any, next any) []any {
	out := make([]any, len(path), len(path)+1)
	copy(out, path)
	out = append(out, next)
	return out
}

func renderPath(path []any) string {
	if len(path) == 0 {
		return "payload"
	}

	var b strings.Builder
	b.WriteString("payload")

	for _, segment := range path {
		switch typed := segment.(type) {
		case int:
			b.WriteString("[")
			b.WriteString(strconv.Itoa(typed))
			b.WriteString("]")
		case string:
			b.WriteString("[")
			b.WriteString(strconv.Quote(typed))
			b.WriteString("]")
		}
	}

	return b.String()
}

func fail(err error, pretty bool) {
	writeJSON(simErrorResponse{
		OK:    false,
		Error: err.Error(),
	}, pretty)
	os.Exit(1)
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
		return nil, errors.New("response kosong atau tidak valid")
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
				return nil, errors.New("payload 'd' kosong")
			}

			var payload any
			if err := json.Unmarshal([]byte(payloadText), &payload); err != nil {
				return nil, fmt.Errorf("decode payload: %w", err)
			}
			return payload, nil
		}

		payload, exists := envelope["d"]
		if !exists {
			return nil, errors.New("field 'd' tidak ditemukan")
		}
		return payload, nil
	case '[':
		var payload any
		if err := json.Unmarshal([]byte(jsonText), &payload); err != nil {
			return nil, fmt.Errorf("decode payload: %w", err)
		}
		return payload, nil
	default:
		return nil, errors.New("format response tidak dikenali")
	}
}

func findPlaceBlocks(payload any) [][]any {
	blocks := make([][]any, 0)
	seen := make(map[string]struct{})

	walk(payload, func(node any) bool {
		items, ok := node.([]any)
		if !ok || len(items) < 120 {
			return true
		}

		dataID, placeID := extractIDs(items)
		if dataID == "" || placeID == "" {
			return true
		}

		key := dataID + "|" + placeID
		if _, exists := seen[key]; exists {
			return true
		}

		seen[key] = struct{}{}
		blocks = append(blocks, items)
		return true
	})

	return blocks
}

func extractIDs(items []any) (dataID string, placeID string) {
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			continue
		}
		if dataID == "" && dataIDPattern.MatchString(text) {
			dataID = text
		}
		if placeID == "" && strings.HasPrefix(text, "ChIJ") {
			placeID = text
		}
		if dataID != "" && placeID != "" {
			break
		}
	}
	return dataID, placeID
}

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
