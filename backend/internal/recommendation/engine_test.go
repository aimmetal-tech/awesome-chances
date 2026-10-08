package recommendation

import (
	"awesome-chances/backend/internal/adapters/demo"
	"awesome-chances/backend/internal/model"
	"testing"
)

func TestRankingRespectsBlockingSkillAndConfidence(t *testing.T) {
	catalog, err := demo.Load()
	if err != nil {
		t.Fatal(err)
	}
	p := catalog.Profile
	p.Skills["web"] = p.Skills["python"] // use a valid estimate, then remove a critical skill
	missing := p.Skills["git"]
	missing.Level = 0
	p.Skills["git"] = missing
	p.Interests["frontend"] = 1
	results := Rank(p, catalog.Tasks, DefaultConfig())
	seenPreparation := false
	for _, result := range results {
		if result.Score < 0 || result.Score > 100 {
			t.Fatalf("invalid score %d", result.Score)
		}
		if !result.LowConfidence {
			t.Error("self-report must remain low confidence")
		}
		if result.Readiness == "prepare" {
			seenPreparation = true
		} else if seenPreparation {
			t.Error("blocked tasks must follow approachable tasks")
		}
		if result.Task.ID == "ui-issue" && result.Readiness != "prepare" {
			t.Error("missing Git cannot be hidden by interest or other skills")
		}
	}
}

func TestInterestChangesOrdering(t *testing.T) {
	catalog, _ := demo.Load()
	p := catalog.Profile
	p.Interests["backend"] = 0
	p.Interests["frontend"] = 1
	p.Interests["ai"] = 0
	p.Interests["security"] = 0
	results := Rank(p, catalog.Tasks, DefaultConfig())
	if results[0].Task.Category != "frontend" {
		t.Fatalf("expected frontend task first, got %s", results[0].Task.ID)
	}
}

func TestModerateChallengeScoresAboveVeryEasyAndVeryHard(t *testing.T) {
	catalog, _ := demo.Load()
	task := catalog.Tasks[3]
	for _, trial := range []struct {
		difficulty float64
		wantBelow  bool
	}{{.35, false}, {.01, true}, {.95, true}} {
		task.Difficulty = trial.difficulty
		got := Rank(catalog.Profile, []model.Task{task}, DefaultConfig())[0].Scores.Difficulty
		if trial.wantBelow && got >= .9 {
			t.Errorf("extreme difficulty should score lower, got %f", got)
		}
		if !trial.wantBelow && got < .99 {
			t.Errorf("moderate challenge should score highly, got %f", got)
		}
	}
}
