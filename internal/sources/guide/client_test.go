package guide

import "testing"

func TestExtractTeamsIncludesMainAndSpares(t *testing.T) {
	slots := []teamSlot{
		{Main: &role{RoleGbID: float64(1103)}, Spares: []role{{RoleGbID: "1104"}}},
		{Main: &role{RoleGbID: "1105"}},
	}
	teams := extractTeams(1102, slots)
	if len(teams) != 2 {
		t.Fatalf("teams = %#v", teams)
	}
	if teams[0][1] != 1103 || teams[1][1] != 1104 {
		t.Fatalf("unexpected teams: %#v", teams)
	}
}

func TestLocalizedTextFallsBackToEnglish(t *testing.T) {
	text := localizedText([]localized{{Language: "en", IntroductionName: "Guide"}}, "pt")
	if text.IntroductionName != "Guide" {
		t.Fatalf("text = %#v", text)
	}
}
