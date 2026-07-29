package assets

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestImageURL(t *testing.T) {
	got, err := ImageURL("/Game/Aki/UI/UIResources/Common/Image/IconRoleHead256/Test.Test")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://static.nanoka.cc/assets/ww/UIResources/Common/Image/IconRoleHead256/Test.webp"
	if got != want {
		t.Fatalf("ImageURL() = %q, want %q", got, want)
	}
}

func TestImageURLSupportsNanokaSkillIcons(t *testing.T) {
	got, err := ImageURL("/Game/Aki/UI/UIResources/Common/Atlas/SkillIcon/SkillIconLuokeke/SP_IconLuokekeC1.SP_IconLuokekeC1")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://static.nanoka.cc/assets/ww/UIResources/Common/Atlas/SkillIcon/SkillIconLuokeke/SP_IconLuokekeC1.webp"
	if got != want {
		t.Fatalf("ImageURL() = %q, want %q", got, want)
	}
}

func TestHandlerOnlyServesCacheRoot(t *testing.T) {
	cache := NewCache(filepath.Join(t.TempDir(), "assets"), nil)
	request := httptest.NewRequest(http.MethodGet, "/not-cache/file.webp", nil)
	response := httptest.NewRecorder()
	cache.Handler().ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestImageURLRejectsTraversal(t *testing.T) {
	_, err := ImageURL("../../secret")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("unexpected error: %v", err)
	}
}
