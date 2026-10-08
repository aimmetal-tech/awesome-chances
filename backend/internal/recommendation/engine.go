package recommendation

import (
	"math"
	"sort"

	"awesome-chances/backend/internal/model"
)

// Config belongs to the backend; UI and BFF never reproduce scoring rules.
type Config struct {
	Weights     model.Scores
	TargetGap   float64
	BlockingGap float64
}

func DefaultConfig() Config {
	return Config{Weights: model.Scores{Skill: .35, Interest: .25, Difficulty: .20, Goal: .10, Quality: .10}, TargetGap: .10, BlockingGap: .25}
}
func clamp(v float64) float64 { return math.Min(1, math.Max(0, v)) }

func Rank(profile model.Profile, tasks []model.Task, cfg Config) []model.Recommendation {
	results := make([]model.Recommendation, 0, len(tasks))
	for _, task := range tasks {
		gaps := make([]model.Gap, 0)
		skill, ability, lowConfidence := 1.0, 0.0, false
		for id, required := range task.RequiredSkills {
			estimate := profile.Skills[id]
			skill = math.Min(skill, clamp(estimate.Level/math.Max(required, .01)))
			ability += estimate.Level
			lowConfidence = lowConfidence || estimate.Confidence < .5
			if gap := required - estimate.Level; gap > .01 {
				gaps = append(gaps, model.Gap{Skill: id, Gap: gap})
			}
		}
		if len(task.RequiredSkills) > 0 {
			ability /= float64(len(task.RequiredSkills))
		}
		sort.Slice(gaps, func(i, j int) bool {
			if gaps[i].Gap == gaps[j].Gap {
				return gaps[i].Skill < gaps[j].Skill
			}
			return gaps[i].Gap > gaps[j].Gap
		})
		goal := .4
		if (profile.Goal == "project" && task.Type == "project") || (profile.Goal == "opensource" && task.Type == "issue") || (profile.Goal == "competition" && task.Type == "competition") {
			goal = 1
		}
		scores := model.Scores{Skill: skill, Interest: profile.Interests[task.Category], Difficulty: clamp(1 - math.Abs(task.Difficulty-ability-cfg.TargetGap)*2), Goal: goal, Quality: task.Quality}
		w := cfg.Weights
		score := int(math.Round((scores.Skill*w.Skill + scores.Interest*w.Interest + scores.Difficulty*w.Difficulty + scores.Goal*w.Goal + scores.Quality*w.Quality) * 100))
		readiness, reason := "ready", "契合你的"+model.InterestLabels[task.Category]+"兴趣，当前基础覆盖了示例任务要求。"
		if len(gaps) > 0 {
			reason = "契合你的" + model.InterestLabels[task.Category] + "兴趣，" + model.SkillLabels[gaps[0].Skill] + "有小幅挑战，可以边学边做。"
			if gaps[0].Gap > cfg.BlockingGap {
				readiness = "prepare"
				reason = "符合你的兴趣，建议先补齐" + model.SkillLabels[gaps[0].Skill] + "基础，再尝试这个任务。"
			}
		}
		results = append(results, model.Recommendation{Task: task, Score: score, Scores: scores, Gaps: gaps, Reason: reason, Readiness: readiness, LowConfidence: lowConfidence, RuleVersion: "demo-baseline-v1"})
	}
	// Preparation suggestions stay visible, after currently approachable tasks.
	// Interest cannot outweigh a missing blocking skill.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Readiness != results[j].Readiness {
			return results[i].Readiness == "ready"
		}
		if results[i].Score == results[j].Score {
			return results[i].Task.ID < results[j].Task.ID
		}
		return results[i].Score > results[j].Score
	})
	return results
}
