package routing

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func resetSuggestCache() {
	suggestCacheMu.Lock()
	suggestCache = map[string][]AddressSuggestion{}
	suggestCacheMu.Unlock()
}

func TestSuggestAddressesShortQuery(t *testing.T) {
	res, err := SuggestAddresses("a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected no suggestions for a short query, got %v", res)
	}
}

func TestSuggestionsCache(t *testing.T) {
	resetSuggestCache()

	if _, ok := cachedSuggestions("amsterdam"); ok {
		t.Fatal("expected cache miss for an unknown query")
	}

	storeSuggestions("Amsterdam ", []AddressSuggestion{{Display: "Amsterdam, Netherlands"}})
	res, ok := cachedSuggestions("amsterdam")
	if !ok {
		t.Fatal("expected cache hit for a case/trim-insensitive variant")
	}
	if len(res) != 1 || res[0].Display != "Amsterdam, Netherlands" {
		t.Fatalf("unexpected cached value: %#v", res)
	}
}

func TestParsePointLL(t *testing.T) {
	lat, lon, ok := parsePointLL("POINT(4.89727884 52.37658789)")
	if !ok {
		t.Fatal("expected a parsed POINT")
	}
	if lat < 52.37 || lat > 52.38 || lon < 4.89 || lon > 4.90 {
		t.Fatalf("unexpected coords lat=%f lon=%f", lat, lon)
	}
	if _, _, ok := parsePointLL(""); ok {
		t.Fatal("empty POINT should not parse")
	}
	if _, _, ok := parsePointLL("POINT(oops)"); ok {
		t.Fatal("invalid POINT should not parse")
	}
}

func TestParsePdokDocs(t *testing.T) {
	body := []byte(`{
		"response": {
			"docs": [
				{"weergavenaam": "Damrak, Amsterdam", "centroide_ll": "POINT(4.8958 52.3751)"},
				{"weergavenaam": "", "centroide_ll": "POINT(4.9 52.3)"},
				{"weergavenaam": "Damrak 18-1, Amsterdam"}
			]
		}
	}`)
	got := parsePdokDocs(body)
	if len(got) != 2 {
		t.Fatalf("got %d docs, want 2", len(got))
	}
	if got[0].Display != "Damrak, Amsterdam" || got[0].Lat == 0 || got[0].Lon == 0 {
		t.Fatalf("unexpected first doc: %#v", got[0])
	}
	if got[1].Display != "Damrak 18-1, Amsterdam" {
		t.Fatalf("unexpected second doc: %#v", got[1])
	}
}

func TestSuggestAddressesUsesPdokAndIsSilentOnError(t *testing.T) {
	resetSuggestCache()
	LoadEnv()

	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if r.URL.Path != "/suggest" {
			t.Errorf("path = %s, want /suggest", r.URL.Path)
		}
		if r.URL.Query().Get("q") != "damrak" {
			t.Errorf("q = %s, want damrak", r.URL.Query().Get("q"))
		}
		io.WriteString(w, `{"response":{"docs":[{"weergavenaam":"Damrak, Amsterdam"}]}}`)
	}))
	defer srv.Close()

	old := pdokURL
	pdokURL = srv.URL
	defer func() { pdokURL = old }()

	res, err := SuggestAddresses("damrak")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res) != 1 || res[0].Display != "Damrak, Amsterdam" {
		t.Fatalf("unexpected suggestions: %#v", res)
	}

	res2, err := SuggestAddresses("Damrak")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res2) != 1 {
		t.Fatalf("expected cached result, got %#v", res2)
	}
	if hits != 1 {
		t.Fatalf("http hits = %d, want 1 (second call cached)", hits)
	}

	fail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusTooManyRequests)
	}))
	defer fail.Close()
	pdokURL = fail.URL
	resetSuggestCache()

	res, err = SuggestAddresses("rotterdam")
	if err != nil {
		t.Fatalf("rate-limit/error must stay silent, got %v", err)
	}
	if res != nil {
		t.Fatalf("expected no suggestions on error, got %#v", res)
	}
}

func TestGeocodePrefersPdok(t *testing.T) {
	LoadEnv()

	pdok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/free" {
			t.Errorf("path = %s, want /free", r.URL.Path)
		}
		io.WriteString(w, `{"response":{"docs":[{"weergavenaam":"Damrak 18-1, Amsterdam","centroide_ll":"POINT(4.89727884 52.37658789)"}]}}`)
	}))
	defer pdok.Close()

	nominatim := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Nominatim should not be called when PDOK succeeds")
	}))
	defer nominatim.Close()

	oldPdok, oldNom := pdokURL, nominatimURL
	pdokURL = pdok.URL
	nominatimURL = nominatim.URL
	defer func() {
		pdokURL = oldPdok
		nominatimURL = oldNom
	}()

	lat, lon, err := Geocode("Damrak 18-1, Amsterdam")
	if err != nil {
		t.Fatalf("geocode: %v", err)
	}
	if lat < 52.37 || lat > 52.38 || lon < 4.89 || lon > 4.90 {
		t.Fatalf("unexpected coords lat=%f lon=%f", lat, lon)
	}
}

func TestPdokMatchesQuery(t *testing.T) {
	if !pdokMatchesQuery("Dam, Amsterdam", "Dam, Amsterdam") {
		t.Fatal("exact city+street must match")
	}
	if pdokMatchesQuery("Centraal Station, Rotterdam", "Metrostation Centraal Station, Amsterdam") {
		t.Fatal("must not accept an Amsterdam hit for a Rotterdam query")
	}
	if pdokMatchesQuery("Centraal Station, Rotterdam", "Centraal Busstation, Breda") {
		t.Fatal("must not accept Breda for Rotterdam")
	}
	if !pdokMatchesQuery("Stationsplein Rotterdam", "Stationsplein, Rotterdam") {
		t.Fatal("token order and punctuation must not matter")
	}
}

func TestGeocodeRejectsWrongCityPdokHit(t *testing.T) {
	LoadEnv()

	pdok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"response":{"docs":[{"weergavenaam":"Metrostation Centraal Station, Amsterdam","centroide_ll":"POINT(4.89985889 52.37816541)"}]}}`)
	}))
	defer pdok.Close()

	nominatim := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `[{"display_name":"Rotterdam Centraal","lat":"51.9250","lon":"4.4692"}]`)
	}))
	defer nominatim.Close()

	oldPdok, oldNom := pdokURL, nominatimURL
	pdokURL = pdok.URL
	nominatimURL = nominatim.URL
	lastNominatim = time.Time{}
	defer func() {
		pdokURL = oldPdok
		nominatimURL = oldNom
	}()

	lat, lon, err := Geocode("Centraal Station, Rotterdam")
	if err != nil {
		t.Fatalf("geocode: %v", err)
	}
	if lat < 51.9 || lat > 51.95 || lon < 4.46 || lon > 4.48 {
		t.Fatalf("expected Rotterdam Centraal, got lat=%f lon=%f", lat, lon)
	}
}

func TestLiveAmsterdamToRotterdamRoute(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test")
	}
	LoadEnv()
	sugs, err := SuggestAddresses("Dam Amsterdam")
	if err != nil {
		t.Skipf("PDOK unavailable: %v", err)
	}
	if len(sugs) == 0 {
		t.Skip("PDOK returned no suggestions")
	}
	route, err := CalculateRoute("Dam, Amsterdam", "Centraal Station, Rotterdam", "fastest")
	if err != nil {
		t.Skipf("live route unavailable: %v", err)
	}
	if route.DistanceKm < 40 || route.DistanceKm > 120 {
		t.Fatalf("unexpected Amsterdam-Rotterdam distance: %.1f km", route.DistanceKm)
	}
	if route.DurationMin < 30 || route.DurationMin > 180 {
		t.Fatalf("unexpected duration: %.0f min", route.DurationMin)
	}
}

func TestGeocodeFallsBackToNominatim(t *testing.T) {
	LoadEnv()

	pdok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer pdok.Close()

	nominatim := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `[{"display_name":"Damrak, Amsterdam","lat":"52.3751","lon":"4.8958"}]`)
	}))
	defer nominatim.Close()

	oldPdok, oldNom := pdokURL, nominatimURL
	pdokURL = pdok.URL
	nominatimURL = nominatim.URL
	lastNominatim = time.Time{}
	defer func() {
		pdokURL = oldPdok
		nominatimURL = oldNom
	}()

	lat, lon, err := Geocode("Damrak, Amsterdam")
	if err != nil {
		t.Fatalf("geocode: %v", err)
	}
	if lat < 52.37 || lat > 52.38 || lon < 4.89 || lon > 4.90 {
		t.Fatalf("unexpected coords lat=%f lon=%f", lat, lon)
	}
}
