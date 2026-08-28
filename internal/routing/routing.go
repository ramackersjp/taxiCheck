package routing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type RouteResult struct {
	DistanceKm  float64
	DurationMin float64
}

const defaultPdokURL = "https://api.pdok.nl/bzk/locatieserver/search/v3_1"

var (
	osrmURL       string
	nominatimURL  string
	pdokURL       string
	userAgent     string
	httpClient    = &http.Client{Timeout: 15 * time.Second}
	loaded        bool
	nominatimMu   sync.Mutex
	lastNominatim time.Time
)

func LoadEnv() {
	if loaded {
		return
	}
	loaded = true

	// Try to find .env next to the executable
	exe, err := os.Executable()
	if err == nil {
		envPath := filepath.Join(filepath.Dir(exe), ".env")
		godotenv.Load(envPath)
	}
	// Also try current directory
	godotenv.Load()

	osrmURL = os.Getenv("OSRM_URL")
	if osrmURL == "" {
		osrmURL = "https://router.project-osrm.org"
	}

	nominatimURL = os.Getenv("NOMINATIM_URL")
	if nominatimURL == "" {
		nominatimURL = "https://nominatim.openstreetmap.org"
	}

	if v := os.Getenv("PDOK_URL"); v != "" {
		pdokURL = v
	} else if pdokURL == "" {
		pdokURL = defaultPdokURL
	}

	userAgent = os.Getenv("USER_AGENT")
	if userAgent == "" {
		userAgent = "TaxiCheck/1.0 (https://github.com/ramackersjp/taxiCheck)"
	}
}

type nominatimResult struct {
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
}

func waitForNominatim() {
	nominatimMu.Lock()
	elapsed := time.Since(lastNominatim)
	if elapsed < time.Second {
		time.Sleep(time.Second - elapsed)
	}
	lastNominatim = time.Now()
	nominatimMu.Unlock()
}

type AddressSuggestion struct {
	Display string
	Lat     float64
	Lon     float64
}

type osrmRoute struct {
	Legs []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"legs"`
}

type osrmResponse struct {
	Code   string      `json:"code"`
	Routes []osrmRoute `json:"routes"`
}

func Geocode(address string) (float64, float64, error) {
	LoadEnv()

	if lat, lon, err := geocodePdok(address); err == nil {
		return lat, lon, nil
	}

	return geocodeNominatim(address)
}

func geocodeNominatim(address string) (float64, float64, error) {
	query := address + ", Netherlands"
	params := url.Values{
		"q":            {query},
		"format":       {"json"},
		"limit":        {"1"},
		"countrycodes": {"nl"},
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		waitForNominatim()

		reqURL := fmt.Sprintf("%s/search?%s", nominatimURL, params.Encode())
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("geocoding request failed: %w", err)
			continue
		}

		if resp.StatusCode == 429 {
			resp.Body.Close()
			lastErr = fmt.Errorf("address lookup temporarily unavailable, try again")
			time.Sleep(2 * time.Second)
			continue
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %w", err)
			continue
		}

		if resp.StatusCode != 200 {
			lastErr = fmt.Errorf("geocoding returned status %d", resp.StatusCode)
			continue
		}

		var results []nominatimResult
		if err := json.Unmarshal(body, &results); err != nil {
			lastErr = fmt.Errorf("failed to parse geocoding response: %w", err)
			continue
		}

		if len(results) == 0 {
			return 0, 0, fmt.Errorf("address not found: %s", address)
		}

		lat, err := strconv.ParseFloat(results[0].Lat, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid latitude: %w", err)
		}
		lon, err := strconv.ParseFloat(results[0].Lon, 64)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid longitude: %w", err)
		}

		return lat, lon, nil
	}

	return 0, 0, lastErr
}

func GetRoute(lat1, lon1, lat2, lon2 float64, mode string) (*RouteResult, error) {
	LoadEnv()

	coords := fmt.Sprintf("%f,%f;%f,%f", lon1, lat1, lon2, lat2)
	reqURL := fmt.Sprintf("%s/route/v1/driving/%s?overview=false&alternatives=true", osrmURL, coords)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create routing request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("routing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("routing returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("failed to read routing response: %w", err)
	}

	var osrmResp osrmResponse
	if err := json.Unmarshal(body, &osrmResp); err != nil {
		return nil, fmt.Errorf("failed to parse routing response: %w", err)
	}

	if osrmResp.Code != "Ok" || len(osrmResp.Routes) == 0 {
		return nil, fmt.Errorf("no route found between the addresses")
	}

	bestRoute := osrmResp.Routes[0]
	if mode == "shortest" && len(osrmResp.Routes) > 1 {
		for _, r := range osrmResp.Routes {
			if len(r.Legs) > 0 && len(bestRoute.Legs) > 0 {
				if r.Legs[0].Distance < bestRoute.Legs[0].Distance {
					bestRoute = r
				}
			}
		}
	}

	if len(bestRoute.Legs) == 0 {
		return nil, fmt.Errorf("route has no legs")
	}

	leg := bestRoute.Legs[0]
	distanceKm := leg.Distance / 1000.0
	durationMin := leg.Duration / 60.0

	return &RouteResult{
		DistanceKm:  distanceKm,
		DurationMin: durationMin,
	}, nil
}

func CalculateRoute(startAddress, endAddress string, mode string) (*RouteResult, error) {
	startAddr := strings.TrimSpace(startAddress)
	endAddr := strings.TrimSpace(endAddress)

	if startAddr == "" || endAddr == "" {
		return nil, fmt.Errorf("both start and end addresses are required")
	}

	lat1, lon1, err := Geocode(startAddr)
	if err != nil {
		return nil, fmt.Errorf("start address: %w", err)
	}

	lat2, lon2, err := Geocode(endAddr)
	if err != nil {
		return nil, fmt.Errorf("end address: %w", err)
	}

	result, err := GetRoute(lat1, lon1, lat2, lon2, mode)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func SuggestAddresses(query string) ([]AddressSuggestion, error) {
	LoadEnv()

	q := strings.TrimSpace(query)
	if len(q) < 2 {
		return nil, nil
	}

	if res, ok := cachedSuggestions(q); ok {
		return res, nil
	}

	params := url.Values{
		"q":    {q},
		"rows": {"10"},
	}

	// PDOK Locatieserver /suggest is built for autocomplete and is not bound
	// by Nominatim's 1 request/second policy, so typing can keep getting
	// addresses. Failures are silent: the TUI keeps the previous list.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	reqURL := fmt.Sprintf("%s/suggest?%s", pdokBase(), params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, nil
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, nil
	}

	suggestions := parsePdokDocs(body)
	storeSuggestions(q, suggestions)
	return suggestions, nil
}

func geocodePdok(address string) (float64, float64, error) {
	q := strings.TrimSpace(address)
	if q == "" {
		return 0, 0, fmt.Errorf("empty address")
	}

	params := url.Values{
		"q":    {q},
		"rows": {"5"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reqURL := fmt.Sprintf("%s/free?%s", pdokBase(), params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("pdok returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, 0, err
	}

	for _, s := range parsePdokDocs(body) {
		if (s.Lat != 0 || s.Lon != 0) && pdokMatchesQuery(q, s.Display) {
			return s.Lat, s.Lon, nil
		}
	}
	return 0, 0, fmt.Errorf("address not found: %s", address)
}

// pdokMatchesQuery requires every meaningful query token to appear in the
// PDOK display name. Without this, "Centraal Station, Rotterdam" matches
// "Metrostation Centraal Station, Amsterdam" and the fare is calculated
// for the wrong city.
func pdokMatchesQuery(query, display string) bool {
	tokens := queryTokens(query)
	if len(tokens) == 0 {
		return true
	}
	d := strings.ToLower(display)
	for _, tok := range tokens {
		if !strings.Contains(d, tok) {
			return false
		}
	}
	return true
}

func queryTokens(q string) []string {
	q = strings.ToLower(strings.ReplaceAll(q, ",", " "))
	var tokens []string
	for _, t := range strings.Fields(q) {
		t = strings.Trim(t, ".-")
		if len(t) < 3 {
			continue
		}
		tokens = append(tokens, t)
	}
	return tokens
}

func pdokBase() string {
	if pdokURL != "" {
		return pdokURL
	}
	return defaultPdokURL
}

type pdokResponse struct {
	Response struct {
		Docs []pdokDoc `json:"docs"`
	} `json:"response"`
}

type pdokDoc struct {
	Weergavenaam string `json:"weergavenaam"`
	CentroideLL  string `json:"centroide_ll"`
}

func parsePdokDocs(body []byte) []AddressSuggestion {
	var parsed pdokResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}
	var suggestions []AddressSuggestion
	for _, d := range parsed.Response.Docs {
		if strings.TrimSpace(d.Weergavenaam) == "" {
			continue
		}
		s := AddressSuggestion{Display: d.Weergavenaam}
		if lat, lon, ok := parsePointLL(d.CentroideLL); ok {
			s.Lat = lat
			s.Lon = lon
		}
		suggestions = append(suggestions, s)
	}
	return suggestions
}

// parsePointLL reads a WKT POINT(lon lat) used by PDOK.
func parsePointLL(s string) (lat, lon float64, ok bool) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "POINT(")
	s = strings.TrimSuffix(s, ")")
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0, 0, false
	}
	lon, err1 := strconv.ParseFloat(parts[0], 64)
	lat, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return lat, lon, true
}

var (
	suggestCacheMu sync.Mutex
	suggestCache   = map[string][]AddressSuggestion{}
)

// cachedSuggestions returns the last fetched results for a query, so repeats
// (e.g. backtracking while typing) are answered instantly without a request.
func cachedSuggestions(q string) ([]AddressSuggestion, bool) {
	suggestCacheMu.Lock()
	defer suggestCacheMu.Unlock()
	res, ok := suggestCache[strings.ToLower(strings.TrimSpace(q))]
	return res, ok
}

// storeSuggestions records results for a query, evicting one entry
// when the cache grows too large.
func storeSuggestions(q string, res []AddressSuggestion) {
	suggestCacheMu.Lock()
	defer suggestCacheMu.Unlock()
	if len(suggestCache) >= 300 {
		for k := range suggestCache {
			delete(suggestCache, k)
			break
		}
	}
	suggestCache[strings.ToLower(strings.TrimSpace(q))] = res
}
