package routing

import (
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

var (
	osrmURL       string
	nominatimURL  string
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

	userAgent = os.Getenv("USER_AGENT")
	if userAgent == "" {
		userAgent = "TaxiCheck/1.0"
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
		nominatimMu.Unlock()
		time.Sleep(time.Second - elapsed)
		nominatimMu.Lock()
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
	waitForNominatim()

	query := address + ", Netherlands"
	params := url.Values{
		"q":            {query},
		"format":       {"json"},
		"limit":        {"1"},
		"countrycodes": {"nl"},
	}

	reqURL := fmt.Sprintf("%s/search?%s", nominatimURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("geocoding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return 0, 0, fmt.Errorf("geocoding returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read response: %w", err)
	}

	var results []nominatimResult
	if err := json.Unmarshal(body, &results); err != nil {
		return 0, 0, fmt.Errorf("failed to parse geocoding response: %w", err)
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

func GetRoute(lat1, lon1, lat2, lon2 float64, mode string) (*RouteResult, error) {
	LoadEnv()

	coords := fmt.Sprintf("%f,%f;%f,%f", lon1, lat1, lon2, lat2)
	reqURL := fmt.Sprintf("%s/route/v1/driving/%s?overview=false&alternatives=true", osrmURL, coords)

	resp, err := httpClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("routing request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("routing returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
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
	waitForNominatim()

	if len(strings.TrimSpace(query)) < 2 {
		return nil, nil
	}

	params := url.Values{
		"q":            {query + ", Netherlands"},
		"format":       {"json"},
		"limit":        {"5"},
		"countrycodes": {"nl"},
	}

	reqURL := fmt.Sprintf("%s/search?%s", nominatimURL, params.Encode())
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []nominatimResult
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	var suggestions []AddressSuggestion
	for _, r := range results {
		lat, err := strconv.ParseFloat(r.Lat, 64)
		if err != nil {
			continue
		}
		lon, err := strconv.ParseFloat(r.Lon, 64)
		if err != nil {
			continue
		}
		suggestions = append(suggestions, AddressSuggestion{
			Display: r.DisplayName,
			Lat:     lat,
			Lon:     lon,
		})
	}

	return suggestions, nil
}
