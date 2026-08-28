package routing

import "testing"

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
	suggestCacheMu.Lock()
	suggestCache = map[string][]AddressSuggestion{}
	suggestCacheMu.Unlock()

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
