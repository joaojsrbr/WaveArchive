package usecase

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wavearchive/internal/domain"
	convenesource "wavearchive/internal/sources/convene"
)

func TestSummarizeConvenePoolUsesIntervalsAfterFirstKnownFiveStar(t *testing.T) {
	pool, _ := convenesource.PoolByType(1)
	pulls := []domain.ConvenePull{
		{ID: 1, PoolType: 1, ItemName: "Verina", Rarity: 5, ObtainedAt: "2026-01-01 10:00:00", SourceIndex: 4},
		{ID: 2, PoolType: 1, ItemName: "Item", Rarity: 3, ObtainedAt: "2026-01-02 10:00:00", SourceIndex: 3},
		{ID: 3, PoolType: 1, ItemName: "Item", Rarity: 4, ObtainedAt: "2026-01-03 10:00:00", SourceIndex: 2},
		{ID: 4, PoolType: 1, ItemName: "Jiyan", Rarity: 5, ObtainedAt: "2026-01-04 10:00:00", SourceIndex: 1},
		{ID: 5, PoolType: 1, ItemName: "Item", Rarity: 3, ObtainedAt: "2026-01-05 10:00:00", SourceIndex: 0},
	}
	summary := summarizeConvenePool(pool.Domain(), pulls)
	if summary.CurrentPity != 1 {
		t.Fatalf("current pity = %d, want 1", summary.CurrentPity)
	}
	if summary.AveragePity5 != 3 {
		t.Fatalf("average pity = %v, want 3", summary.AveragePity5)
	}
	if summary.GuaranteeState != "not_guaranteed" {
		t.Fatalf("guarantee = %q, want not_guaranteed", summary.GuaranteeState)
	}
	if summary.HistoryPartial {
		t.Fatal("history after a known five-star should prove current pity")
	}
}

func TestSummarizeConvenePoolMarksUnprovenPityAsPartial(t *testing.T) {
	pool, _ := convenesource.PoolByType(1)
	summary := summarizeConvenePool(pool.Domain(), []domain.ConvenePull{
		{ID: 1, PoolType: 1, ItemName: "Item", Rarity: 3, ObtainedAt: "2026-01-01 10:00:00"},
	})
	if !summary.HistoryPartial {
		t.Fatal("history without a five-star must be marked partial")
	}
	if summary.CurrentPity != 1 {
		t.Fatalf("current pity = %d, want minimum 1", summary.CurrentPity)
	}
}

func TestBuildConveneOverviewReturnsEmptyCollections(t *testing.T) {
	overview := buildConveneOverview(nil, nil, nil)
	if overview.Pulls == nil {
		t.Fatal("pulls must be an empty collection, not null")
	}
	if overview.Pools == nil {
		t.Fatal("pools must be an empty collection, not null")
	}
}

func TestExtractConveneURLDecodesObfuscatedClientLog(t *testing.T) {
	const expected = "https://aki-gm-resources-oversea.aki-game.net/aki/gacha/index.html#/record?svr_id=6&player_id=500000001&record_id=temporary"
	plain := []byte("prefix " + expected + " suffix")
	encoded := encodeConveneClientLogForTest(plain)
	path := filepath.Join(t.TempDir(), "Client.log")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}

	actual, err := extractConveneURL(path)
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("URL = %q, want %q", actual, expected)
	}
}

func TestExtractConveneURLReadsPlainDebugLog(t *testing.T) {
	const expected = "https://aki-gm-resources-oversea.aki-game.net/aki/gacha/index.html#/record?svr_id=6&player_id=500000001&record_id=temporary"
	path := filepath.Join(t.TempDir(), "debug.log")
	if err := os.WriteFile(path, []byte("first\n"+expected+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	actual, err := extractConveneURL(path)
	if err != nil {
		t.Fatal(err)
	}
	if actual != expected {
		t.Fatalf("URL = %q, want %q", actual, expected)
	}
}

func TestExistingConveneLogsUsesNewestFirstAndDeduplicates(t *testing.T) {
	root := t.TempDir()
	older := filepath.Join(root, "Client.log")
	newer := filepath.Join(root, "debug.log")
	if err := os.WriteFile(older, []byte("older"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newer, []byte("newer"), 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if err := os.Chtimes(older, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newer, now, now); err != nil {
		t.Fatal(err)
	}

	logs := existingConveneLogs([]string{older, newer, older})
	if len(logs) != 2 {
		t.Fatalf("logs = %d, want 2", len(logs))
	}
	if logs[0].path != newer {
		t.Fatalf("first log = %q, want newest %q", logs[0].path, newer)
	}
}

func TestExtractConveneURLFromConfiguredLog(t *testing.T) {
	path := os.Getenv("WAVEARCHIVE_CONVENE_LOG")
	if path == "" {
		t.Skip("WAVEARCHIVE_CONVENE_LOG is not configured")
	}
	rawURL, err := extractConveneURL(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rawURL, "https://aki-gm-resources") {
		t.Fatal("configured log returned an unexpected Convene URL")
	}
}

func encodeConveneClientLogForTest(content []byte) []byte {
	encoded := make([]byte, len(content))
	for index, value := range content {
		if value&1 == 0 {
			encoded[index] = value ^ 0xA5
		} else {
			encoded[index] = value ^ 0xEF
		}
	}
	return encoded
}
