package gmapsparser

import (
	"fmt"
	"math/big"
	"strings"
)

func extractOperatingHours(hoursBlock any) map[string]string {
	weekRows, ok := asSlice(safeGet(hoursBlock, 0))
	if !ok {
		return map[string]string{}
	}

	result := make(map[string]string)
	for _, dayRow := range weekRows {
		row, ok := asSlice(dayRow)
		if !ok || len(row) < 4 {
			continue
		}

		dayName, ok := asString(row[0])
		if !ok || strings.TrimSpace(dayName) == "" {
			continue
		}

		normalizedDay := strings.ToLower(strings.TrimSpace(dayName))
		dayKey := dayMap[normalizedDay]
		if dayKey == "" {
			dayKey = normalizedDay
		}

		slotText := ""
		slots, ok := asSlice(row[3])
		if ok && len(slots) > 0 {
			firstSlot, ok := asSlice(slots[0])
			if ok && len(firstSlot) > 0 {
				if text, ok := asString(firstSlot[0]); ok {
					slotText = text
				}
			}
		}

		result[dayKey] = slotText
	}

	return result
}

func extractOpenState(hoursBlock any) *string {
	stateBlock, ok := asSlice(safeGet(hoursBlock, 1, 4))
	if !ok || len(stateBlock) == 0 {
		return nil
	}
	return stringOrNil(stateBlock[0])
}

func extractExtensions(bundle any, wantedEnabled bool) []map[string][]string {
	// Prioritaskan grouped categories seperti service_options/highlights.
	grouped := extractExtensionsByKeyFilter(bundle, wantedEnabled, isExtensionGroupedKey)
	if len(grouped) > 0 {
		return grouped
	}

	// Fallback untuk payload yang hanya berisi key mentah /geo/type/...
	return extractExtensionsByKeyFilter(bundle, wantedEnabled, func(_ string) bool { return true })
}

func isExtensionGroupedKey(key string) bool {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return false
	}

	return !strings.HasPrefix(trimmed, "/geo/type/")
}

func extractExtensionsByKeyFilter(bundle any, wantedEnabled bool, includeKey func(string) bool) []map[string][]string {
	sections, ok := asSlice(bundle)
	if !ok {
		return []map[string][]string{}
	}

	type groupBucket struct {
		key    string
		labels []string
		seen   map[string]struct{}
	}

	orderedKeys := make([]string, 0)
	buckets := make(map[string]*groupBucket)

	for sectionIndex, section := range sections {
		groups, ok := asSlice(section)
		if !ok {
			continue
		}

		for _, group := range groups {
			row, ok := asSlice(group)
			if !ok || len(row) < 3 {
				continue
			}

			key, ok := asString(row[0])
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			if key == "" || !includeKey(key) {
				continue
			}

			items, ok := asSlice(row[2])
			if !ok {
				continue
			}

			bucket, exists := buckets[key]
			if !exists {
				bucket = &groupBucket{key: key, labels: make([]string, 0), seen: make(map[string]struct{})}
				buckets[key] = bucket
				orderedKeys = append(orderedKeys, key)
			}

			for _, item := range items {
				entry, ok := asSlice(item)
				if !ok || len(entry) < 2 {
					continue
				}

				label, ok := asString(entry[1])
				if !ok {
					continue
				}
				label = strings.TrimSpace(label)
				if label == "" {
					continue
				}

				enabled, known := parseExtensionEnabled(entry)
				if !known {
					enabled = inferEnabledFromSection(sectionIndex)
				}
				if enabled != wantedEnabled {
					continue
				}

				normalized := strings.ToLower(label)
				if _, seen := bucket.seen[normalized]; seen {
					continue
				}

				bucket.seen[normalized] = struct{}{}
				bucket.labels = append(bucket.labels, label)
			}
		}
	}

	output := make([]map[string][]string, 0)
	for _, key := range orderedKeys {
		bucket := buckets[key]
		if bucket == nil || len(bucket.labels) == 0 {
			continue
		}
		output = append(output, map[string][]string{key: bucket.labels})
	}

	return output
}

func parseExtensionEnabled(entry []any) (bool, bool) {
	// Flag status di index 3 lebih stabil untuk membedakan supported vs unsupported.
	status := safeGet(entry, 3)
	if status == nil {
		return true, true
	}
	if value, ok := asInt(status); ok {
		return value != 1, true
	}

	if value, ok := asInt(safeGet(entry, 2, 0)); ok {
		return value == 1, true
	}
	if value, ok := asInt(safeGet(entry, 2, 2, 0)); ok {
		return value == 1, true
	}
	return false, false
}

func inferEnabledFromSection(sectionIndex int) bool {
	if sectionIndex == 5 {
		return false
	}
	return true
}

func extractServiceOptions(bundle any) map[string]bool {
	options := map[string]bool{
		"dine_in":  false,
		"takeout":  false,
		"delivery": false,
	}

	sections, ok := asSlice(bundle)
	if !ok {
		return options
	}

	for sectionIndex, section := range sections {
		groups, ok := asSlice(section)
		if !ok {
			continue
		}

		for _, group := range groups {
			row, ok := asSlice(group)
			if !ok || len(row) < 3 {
				continue
			}

			items, ok := asSlice(row[2])
			if !ok {
				continue
			}

			for _, item := range items {
				entry, ok := asSlice(item)
				if !ok || len(entry) == 0 {
					continue
				}

				featureID, ok := asString(entry[0])
				if !ok {
					continue
				}

				enabled, known := parseExtensionEnabled(entry)
				if !known {
					enabled = inferEnabledFromSection(sectionIndex)
				}
				if !enabled {
					continue
				}

				switch {
				case strings.Contains(featureID, "serves_dine_in"):
					options["dine_in"] = true
				case strings.Contains(featureID, "has_takeout"):
					options["takeout"] = true
				case strings.Contains(featureID, "has_delivery") || strings.Contains(featureID, "no_contact_delivery"):
					options["delivery"] = true
				}
			}
		}
	}

	return options
}

func extractPhone(block []any) *string {
	if text := stringOrNil(safeGet(block, 178, 0, 0)); text != nil {
		return text
	}

	var phone *string
	walk(block, func(node any) bool {
		if phone != nil {
			return false
		}

		text, ok := asString(node)
		if !ok {
			return true
		}

		text = strings.TrimSpace(text)
		if phonePattern.MatchString(text) {
			copy := text
			phone = &copy
			return false
		}
		return true
	})

	return phone
}

func extractWebsite(block []any) *string {
	direct, ok := asSlice(safeGet(block, 7))
	if ok && len(direct) > 0 {
		if website := stringOrNil(direct[0]); website != nil && (strings.HasPrefix(*website, "http://") || strings.HasPrefix(*website, "https://")) {
			return website
		}
	}

	return findFirstURL(block, func(rawURL string) bool {
		return !isGoogleHost(rawURL)
	})
}

func extractThumbnail(block []any) *string {
	if thumb := stringOrNil(safeGet(block, 37, 0, 0, 6, 0)); thumb != nil {
		return thumb
	}

	return findFirstURL(block, func(rawURL string) bool {
		return strings.Contains(rawURL, "googleusercontent.com") || strings.Contains(rawURL, "streetviewpixels-pa.googleapis.com")
	})
}

func extractOrderOnline(block []any) *string {
	return findFirstURL(block, func(rawURL string) bool {
		return strings.Contains(rawURL, "chooseprovider")
	})
}

func extractReviewsLink(block []any) *string {
	if link := stringOrNil(safeGet(block, 4, 3, 0)); link != nil {
		return link
	}

	return findFirstURL(block, func(rawURL string) bool {
		return strings.Contains(rawURL, "reviews?placeid=")
	})
}

func extractReviewSnippets(block []any) []string {
	candidates, ok := asSlice(safeGet(block, 31, 1))
	if !ok {
		return []string{}
	}

	seen := make(map[string]struct{})
	snippets := make([]string, 0)
	for _, item := range candidates {
		entry, ok := asSlice(item)
		if !ok {
			continue
		}

		text := normalizeText(safeGet(entry, 1))
		if text == nil {
			continue
		}

		key := strings.ToLower(*text)
		if _, exists := seen[key]; exists {
			continue
		}

		seen[key] = struct{}{}
		snippets = append(snippets, *text)
	}

	return snippets
}

func looksLikeReviewCore(node any) bool {
	items, ok := asSlice(node)
	if !ok || len(items) < 2 {
		return false
	}

	if _, ok := asString(items[0]); !ok {
		return false
	}
	if _, ok := asSlice(items[1]); !ok {
		return false
	}

	author := stringOrNil(safeGet(items, 1, 4, 5, 0))
	relativeTime := stringOrNil(safeGet(items, 1, 6))
	return author != nil && relativeTime != nil
}

func extractUserReviews(block []any, limit int) []map[string]any {
	reviewContainer, ok := asSlice(safeGet(block, 175))
	if !ok {
		return []map[string]any{}
	}

	reviewCores := make([][]any, 0)
	seenIDs := make(map[string]struct{})
	walk(reviewContainer, func(node any) bool {
		if !looksLikeReviewCore(node) {
			return true
		}

		items, _ := asSlice(node)
		reviewID, ok := asString(safeGet(items, 0))
		if !ok {
			return true
		}
		if _, exists := seenIDs[reviewID]; exists {
			return true
		}

		seenIDs[reviewID] = struct{}{}
		reviewCores = append(reviewCores, items)
		return true
	})

	if len(reviewCores) > limit {
		reviewCores = reviewCores[:limit]
	}

	reviews := make([]map[string]any, 0, len(reviewCores))
	for _, core := range reviewCores {
		photos := make([]string, 0)
		if mediaItems, ok := asSlice(safeGet(core, 2, 2)); ok {
			for _, media := range mediaItems {
				if mediaURL, ok := asString(safeGet(media, 1, 6, 0)); ok {
					photos = append(photos, mediaURL)
				}
			}
		}

		reviews = append(reviews, map[string]any{
			"review_id":     safeGet(core, 0),
			"rating":        safeGet(core, 1, 13, 4),
			"text":          normalizeText(safeGet(core, 2, 15, 0, 0)),
			"relative_time": safeGet(core, 1, 6),
			"review_link":   safeGet(core, 4, 3, 0),
			"author": map[string]any{
				"name":         safeGet(core, 1, 4, 5, 0),
				"id":           safeGet(core, 1, 4, 5, 3),
				"profile_link": safeGet(core, 1, 4, 5, 2, 0),
				"avatar":       safeGet(core, 1, 4, 5, 1),
				"badge":        safeGet(core, 1, 4, 5, 10, 0),
				"reviews":      safeGet(core, 1, 4, 5, 5),
				"photos":       safeGet(core, 1, 4, 5, 6),
			},
			"photos": photos,
		})
	}

	return reviews
}

func extractPhoneDetails(block []any) map[string]any {
	phoneBlock, ok := asSlice(safeGet(block, 178))
	if !ok || len(phoneBlock) == 0 {
		return nil
	}

	item, ok := asSlice(phoneBlock[0])
	if !ok {
		return nil
	}

	details := map[string]any{
		"display":       safeGet(item, 0),
		"international": safeGet(item, 1, 1, 0),
		"normalized":    safeGet(item, 3),
		"tel_uri":       safeGet(item, 5, 0),
	}

	for _, value := range details {
		if stringOrNil(value) != nil {
			return details
		}
	}

	return nil
}

func extractAddressDetails(block []any) map[string]any {
	components := safeGet(block, 2)
	if components == nil {
		components = []any{}
	}

	details := map[string]any{
		"components":   components,
		"neighborhood": safeGet(block, 14),
		"district":     nil,
		"street":       nil,
		"city":         nil,
		"postal_code":  nil,
		"state":        nil,
		"country_code": nil,
		"plus_code": map[string]any{
			"global_code":   nil,
			"compound_code": nil,
		},
	}

	if parts, ok := asSlice(safeGet(block, 183, 1)); ok {
		details["district"] = safeGet(parts, 0)
		details["street"] = safeGet(parts, 1)
		details["city"] = safeGet(parts, 3)
		details["postal_code"] = safeGet(parts, 4)
		details["state"] = safeGet(parts, 5)
		details["country_code"] = safeGet(parts, 6)
	}

	if plus, ok := asSlice(safeGet(block, 183, 2)); ok {
		details["plus_code"] = map[string]any{
			"global_code":   safeGet(plus, 1, 0),
			"compound_code": safeGet(plus, 2, 0),
		}
	}

	if details["country_code"] == nil {
		details["country_code"] = safeGet(block, 243)
	}

	return details
}

func extractPlacePhotos(block []any, limit int) map[string]any {
	photoSections := []int{37, 51, 72, 105, 171}
	seenURLs := make(map[string]struct{})
	photos := make([]map[string]any, 0)

	addPhotoNode := func(node any) {
		if len(photos) >= limit {
			return
		}

		urlValue, ok := asString(safeGet(node, 6, 0))
		if !ok {
			return
		}

		if !strings.Contains(urlValue, "googleusercontent.com") && !strings.Contains(urlValue, "streetviewpixels-pa.googleapis.com") {
			return
		}

		if _, exists := seenURLs[urlValue]; exists {
			return
		}
		seenURLs[urlValue] = struct{}{}

		photo := map[string]any{
			"photo_id":       safeGet(node, 0),
			"type":           safeGet(node, 19),
			"caption":        normalizeText(safeGet(node, 6, 1)),
			"url":            urlValue,
			"size":           nil,
			"thumbnail_size": nil,
		}

		if photo["caption"] == nil {
			photo["caption"] = normalizeText(safeGet(node, 3))
		}

		if size, ok := asSlice(safeGet(node, 6, 2)); ok && len(size) >= 2 {
			photo["size"] = map[string]any{"width": size[0], "height": size[1]}
		}
		if thumbSize, ok := asSlice(safeGet(node, 6, 3)); ok && len(thumbSize) >= 2 {
			photo["thumbnail_size"] = map[string]any{"width": thumbSize[0], "height": thumbSize[1]}
		}

		photos = append(photos, photo)
	}

	for _, sectionIndex := range photoSections {
		section := safeGet(block, sectionIndex)
		if section == nil {
			continue
		}

		walk(section, func(node any) bool {
			addPhotoNode(node)
			return len(photos) < limit
		})

		if len(photos) >= limit {
			break
		}
	}

	count, ok := asInt(safeGet(block, 37, 1))
	if !ok {
		count = len(photos)
	}

	return map[string]any{
		"count": count,
		"items": photos,
	}
}

func extractPopularTimes(block []any) map[string]any {
	busyBlock, ok := asSlice(safeGet(block, 84))
	if !ok {
		return nil
	}

	dayNames := map[int]string{
		1: "monday",
		2: "tuesday",
		3: "wednesday",
		4: "thursday",
		5: "friday",
		6: "saturday",
		7: "sunday",
	}

	dayEntries := busyBlock
	if first, ok := asSlice(safeGet(busyBlock, 0)); ok && len(first) > 0 {
		if _, nested := asSlice(first[0]); nested {
			dayEntries = first
		}
	}

	days := make(map[string]any)
	var currentWait any

	for _, dayEntry := range dayEntries {
		dayNum, ok := asInt(safeGet(dayEntry, 0))
		if !ok {
			continue
		}

		rows, ok := asSlice(safeGet(dayEntry, 1))
		if !ok {
			continue
		}

		dayKey := dayNames[dayNum]
		if dayKey == "" {
			dayKey = fmt.Sprintf("day_%d", dayNum)
		}

		hourly := make([]map[string]any, 0)
		for _, row := range rows {
			hour, ok := asInt(safeGet(row, 0))
			if !ok {
				continue
			}

			hourly = append(hourly, map[string]any{
				"hour":         hour,
				"intensity":    safeGet(row, 1),
				"description":  normalizeText(safeGet(row, 2)),
				"display_time": normalizeText(safeGet(row, 4)),
				"wait":         normalizeText(safeGet(row, 5)),
			})
		}

		days[dayKey] = hourly
		if currentWait == nil {
			currentWait = stringOrNil(safeGet(dayEntry, 3, 0))
		}
	}

	if currentWait == nil {
		currentWait = stringOrNil(safeGet(busyBlock, 1, 3, 0))
	}

	if len(days) == 0 {
		return nil
	}

	return map[string]any{
		"days":         days,
		"current_wait": currentWait,
	}
}

func extractWaitEstimates(block []any) []map[string]any {
	routeBlock, ok := asSlice(safeGet(block, 229, 1))
	if !ok {
		return []map[string]any{}
	}

	modeMap := map[int]string{
		0: "walking",
		1: "transit",
		2: "driving",
		3: "cycling",
	}

	estimates := make([]map[string]any, 0)
	for _, item := range routeBlock {
		modeCode, ok := asInt(safeGet(item, 0, 0))
		if !ok {
			continue
		}

		mode := modeMap[modeCode]
		if mode == "" {
			mode = fmt.Sprintf("mode_%d", modeCode)
		}

		estimates = append(estimates, map[string]any{
			"mode":             mode,
			"duration_seconds": safeGet(item, 0, 1, 0),
			"duration_label":   safeGet(item, 0, 1, 1),
			"target_coordinates": map[string]any{
				"latitude":  safeGet(item, 1, 2),
				"longitude": safeGet(item, 1, 3),
			},
			"score": safeGet(item, 4),
		})
	}

	return estimates
}

func toPlaceOutput(block []any, position int, hl string) map[string]any {
	dataID := safeGet(block, 10)
	title := safeGet(block, 11)
	placeID := safeGet(block, 78)
	providerID := safeGet(block, 89)

	ratingBlock, ratingBlockOK := asSlice(safeGet(block, 4))
	var price any
	var rating any
	var reviews any
	if ratingBlockOK {
		price = safeGet(ratingBlock, 2)
		rating = safeGet(ratingBlock, 7)
		reviews = safeGet(ratingBlock, 8)
	}

	gps, gpsOK := asSlice(safeGet(block, 9))
	var latitude any
	var longitude any
	if gpsOK {
		latitude = safeGet(gps, 2)
		longitude = safeGet(gps, 3)
	}

	types := make([]string, 0)
	for _, value := range toStringSlice(safeGet(block, 13)) {
		types = append(types, value)
	}
	var typeValue any
	if len(types) > 0 {
		typeValue = types[0]
	}

	typeRows, _ := asSlice(safeGet(block, 76))
	typeIDs := make([]string, 0)
	for _, row := range typeRows {
		if typeID, ok := asString(safeGet(row, 0)); ok {
			typeIDs = append(typeIDs, typeID)
		}
	}

	var typeID any
	if len(typeIDs) > 0 {
		typeID = typeIDs[0]
	}

	address := safeGet(block, 39)
	if _, ok := asString(address); !ok {
		address = safeGet(block, 18)
	}

	openState := extractOpenState(safeGet(block, 203))
	operatingHours := extractOperatingHours(safeGet(block, 203))
	extBundle := safeGet(block, 100)
	extensions := extractExtensions(extBundle, true)
	unsupportedExtensions := extractExtensions(extBundle, false)
	serviceOptions := extractServiceOptions(extBundle)
	phone := extractPhone(block)
	phoneDetails := extractPhoneDetails(block)
	website := extractWebsite(block)
	orderOnline := extractOrderOnline(block)
	thumbnail := extractThumbnail(block)
	reviewsLink := extractReviewsLink(block)
	reviewSnippets := extractReviewSnippets(block)
	userReviews := extractUserReviews(block, 20)
	addressDetails := extractAddressDetails(block)
	popularTimes := extractPopularTimes(block)
	waitEstimates := extractWaitEstimates(block)
	timezone := safeGet(block, 30)
	language := safeGet(block, 110)
	region := safeGet(block, 111)

	var dataCID any
	if value, ok := asString(dataID); ok && strings.Contains(value, ":") {
		tail := strings.SplitN(value, ":", 2)[1]
		if strings.HasPrefix(tail, "0x") {
			if bigValue, success := new(big.Int).SetString(tail, 0); success {
				dataCID = bigValue.Text(10)
			}
		}
	}

	return map[string]any{
		"position":               position,
		"title":                  title,
		"place_id":               placeID,
		"data_id":                dataID,
		"data_cid":               dataCID,
		"reviews_link":           reviewsLink,
		"gps_coordinates":        map[string]any{"latitude": latitude, "longitude": longitude},
		"provider_id":            providerID,
		"rating":                 rating,
		"reviews":                reviews,
		"review_snippets":        reviewSnippets,
		"user_reviews":           userReviews,
		"user_reviews_count":     len(userReviews),
		"price":                  price,
		"type":                   typeValue,
		"types":                  types,
		"type_id":                typeID,
		"type_ids":               typeIDs,
		"address":                address,
		"address_details":        addressDetails,
		"open_state":             openState,
		"hours":                  openState,
		"operating_hours":        operatingHours,
		"popular_times":          popularTimes,
		"wait_estimates":         waitEstimates,
		"phone":                  phone,
		"phone_details":          phoneDetails,
		"website":                website,
		"timezone":               timezone,
		"language":               language,
		"region":                 region,
		"extensions":             extensions,
		"unsupported_extensions": unsupportedExtensions,
		"service_options":        serviceOptions,
		"order_online":           orderOnline,
		"thumbnail":              thumbnail,
	}
}
