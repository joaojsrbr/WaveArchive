package convene

import (
	"encoding/json"
	"testing"
)

func TestParseHistoryURLAcceptsOfficialGlobalURL(t *testing.T) {
	raw := "https://aki-gm-resources-oversea.aki-game.net/aki/gacha/index.html#/record?player_id=500123456&svr_id=6&resources_id=abc&record_id=temporary-token&lang=pt"
	credentials, err := ParseHistoryURL(raw)
	if err != nil {
		t.Fatal(err)
	}
	if credentials.PlayerID != "500123456" || credentials.ServerID != "6" {
		t.Fatalf("unexpected credentials: %+v", credentials)
	}
	if credentials.Endpoint != "https://gmserver-api.aki-game2.net/gacha/record/query" {
		t.Fatalf("unexpected endpoint: %s", credentials.Endpoint)
	}
}

func TestParseHistoryURLRejectsUntrustedHost(t *testing.T) {
	_, err := ParseHistoryURL("https://example.com/aki/gacha/index.html#/record?player_id=1&svr_id=2&resources_id=3&record_id=4")
	if err == nil {
		t.Fatal("expected untrusted host to be rejected")
	}
}

func TestPoolsFromSelectListKeepsOfficialKeysAndAddsNewOnes(t *testing.T) {
	pools := PoolsFromSelectList(map[string]string{
		"characterEvent": "Featured Resonator Convene",
		"futureBanner":   "Future Official Convene",
	})
	if len(pools) != 2 {
		t.Fatalf("pool count = %d, want 2", len(pools))
	}
	if pools[0].Type != 1 || pools[0].LocaleKey != "characterEvent" {
		t.Fatalf("known pool mapping changed: %+v", pools[0])
	}
	if pools[1].Type != 14 || pools[1].LocaleKey != "futureBanner" {
		t.Fatalf("dynamic pool not assigned after known types: %+v", pools[1])
	}
}

func TestStringOrNumberAcceptsResourceIDFormats(t *testing.T) {
	for _, test := range []struct {
		name string
		raw  string
		want string
	}{
		{name: "string", raw: `"21020043"`, want: "21020043"},
		{name: "number", raw: `21020043`, want: "21020043"},
		{name: "null", raw: `null`, want: ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			var value stringOrNumber
			if err := json.Unmarshal([]byte(test.raw), &value); err != nil {
				t.Fatal(err)
			}
			if string(value) != test.want {
				t.Fatalf("value = %q, want %q", value, test.want)
			}
		})
	}
}

func TestNormalizeResourceTypeAcceptsPortugueseAPIValues(t *testing.T) {
	if got := normalizeResourceType("Ressonante"); got != "character" {
		t.Fatalf("Ressonante = %q, want character", got)
	}
	if got := normalizeResourceType("Arma"); got != "weapon" {
		t.Fatalf("Arma = %q, want weapon", got)
	}
}
