package nanoka

import (
	"encoding/json"
	"testing"
)

func TestOrderedCharacterIndexPreservesJSONOrder(t *testing.T) {
	var index orderedCharacterIndex
	payload := []byte(`{
		"1606":{"en":"Roccia"},
		"1102":{"en":"Sanhua"},
		"1507":{"en":"Zani"}
	}`)
	if err := json.Unmarshal(payload, &index); err != nil {
		t.Fatal(err)
	}
	if index["1606"].APIOrder != 0 || index["1102"].APIOrder != 1 || index["1507"].APIOrder != 2 {
		t.Fatalf("unexpected API order: %#v", index)
	}
}
