package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"rasanta/internal/service"

	"github.com/gofiber/fiber/v2"
)

type PlaceHandler struct {
	service service.PlaceService
}

func NewPlaceHandler(service service.PlaceService) *PlaceHandler {
	return &PlaceHandler{service: service}
}

func (h *PlaceHandler) Fetch(c *fiber.Ctx) error {
	query := c.Query("query", "bali")
	hl := c.Query("hl", "id")
	gl := c.Query("gl", "id")
	authUser := c.Query("authuser", "0")


	lat, err := parseFloatQuery(c.Query("lat"), -7.2575)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	lng, err := parseFloatQuery(c.Query("lng"), 112.7521)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	statusCode, places, err := h.service.FetchSearch(requestContext(c), service.SearchParams{
		Query:    query,
		Lat:      lat,
		Lng:      lng,
		HL:       hl,
		GL:       gl,
		AuthUser: authUser,
	})
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":       err.Error(),
			"status_code": statusCode,
		})
	}

	return c.JSON(fiber.Map{
		"status_code": statusCode,
		"count":       len(places),
		"data":        places,
	})
}

func (h *PlaceHandler) Detail(c *fiber.Ctx) error {
	pb, err := resolveDetailPB(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	q := c.Query("q")
	if q == "" {
		q = c.Query("query")
	}

	hl := c.Query("hl", "id")
	gl := c.Query("gl", "id")
	authUser := c.Query("authuser", "0")

	statusCode, placeData, err := h.service.FetchDetail(requestContext(c), service.DetailParams{
		PB:       pb,
		Q:        q,
		HL:       hl,
		GL:       gl,
		AuthUser: authUser,
	})
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error":       err.Error(),
			"status_code": statusCode,
		})
	}

	return c.JSON(fiber.Map{
		"status_code": statusCode,
		"data":        placeData,
	})
}

func requestContext(c *fiber.Ctx) context.Context {
	ctx := c.UserContext()
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func parseFloatQuery(raw string, fallback float64) (float64, error) {
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float value %q", raw)
	}
	return value, nil
}

func parseIntQuery(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid int value %q", raw)
	}

	return value, nil
}

func resolveDetailPB(c *fiber.Ctx) (string, error) {
	pb := strings.TrimSpace(c.Query("pb"))
	if pb != "" {
		// Kalau pb dikirim langsung dari client, pakai apa adanya.
		return pb, nil
	}

	dataID := strings.TrimSpace(c.Query("data_id"))
	if dataID == "" {
		return "", fmt.Errorf("data_id, lat, dan long wajib diisi")
	}

	lat, err := normalizeCoordinate("lat", c.Query("lat"))
	if err != nil {
		return "", err
	}

	lngRaw := c.Query("long")
	if strings.TrimSpace(lngRaw) == "" {
		lngRaw = c.Query("lng")
	}
	if strings.TrimSpace(lngRaw) == "" {
		lngRaw = c.Query("longitude")
	}

	lng, err := normalizeCoordinate("long", lngRaw)
	if err != nil {
		return "", err
	}

	return buildDetailPB(dataID, lng, lat), nil
}

func normalizeCoordinate(name, raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", fmt.Errorf("%s wajib diisi saat menggunakan data_id", name)
	}

	value, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return "", fmt.Errorf("%s tidak valid: %q", name, trimmed)
	}

	return strconv.FormatFloat(value, 'f', -1, 64), nil
}

func buildDetailPB(dataID, lng, lat string) string {
	// Format PB untuk detail tempat di Google Maps, berdasarkan observasi dari payload asli.
	return fmt.Sprintf("!1m14!1s%s!3m12!1m3!1d15917.57438272561!2d%s!3d%s!2m3!1f0!2f0!3f0!3m2!1i664!2i772!4f13.1!12m4!2m3!1i360!2i120!4i8!13m57!2m2!1i203!2i100!3m2!2i4!5b1!6m6!1m2!1i86!2i86!1m2!1i408!2i240!7m33!1m3!1e1!2b0!3e3!1m3!1e2!2b1!3e2!1m3!1e2!2b0!3e3!1m3!1e8!2b0!3e3!1m3!1e10!2b0!3e3!1m3!1e10!2b1!3e2!1m3!1e10!2b0!3e4!1m3!1e9!2b1!3e2!2b1!9b0!15m8!1m7!1m2!1m1!1e2!2m2!1i195!2i195!3i20!14m5!1sJiK2aduNOISdseMP0MX4-AU:39!2s1i:0,t:150714,p:JiK2aduNOISdseMP0MX4-AU:39!7e81!12e3!17sJiK2aduNOISdseMP0MX4-AU:45!15m110!1m28!13m9!2b1!3b1!4b1!6i1!8b1!9b1!14b1!20b1!25b1!18m17!3b1!4b1!5b1!6b1!9b1!13b1!14b1!17b1!20b1!21b1!22b1!30b1!32b1!33m1!1b1!34b1!36e2!10m1!8e3!11m1!3e1!17b1!20m2!1e3!1e6!24b1!25b1!26b1!27b1!29b1!30m1!2b1!36b1!37b1!39m3!2m2!2i1!3i1!43b1!52b1!54m1!1b1!55b1!56m1!1b1!61m2!1m1!1e1!65m5!3m4!1m3!1m2!1i224!2i298!72m22!1m8!2b1!5b1!7b1!12m4!1b1!2b1!4m1!1e1!4b1!8m10!1m6!4m1!1e1!4m1!1e3!4m1!1e4!3sother_user_google_review_posts__and__hotel_and_vr_partner_review_posts!6m1!1e1!9b1!89b1!90m2!1m1!1e2!98m3!1b1!2b1!3b1!103b1!113b1!114m3!1b1!2m1!1b1!117b1!122m1!1b1!126b1!127b1!128m1!1b0!21m0!22m1!1e81!29m0!30m6!3b1!6m1!2b1!7m1!2b1!9b1!34m5!7b1!10b1!14b1!15m1!1b0!37i770!39s", dataID, lng, lat)
}
