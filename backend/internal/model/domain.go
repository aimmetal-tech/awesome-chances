package model

import (
	"errors"
	"math"
	"strings"
	"unicode/utf8"
)

type SkillEstimate struct {
	Level      float64 `json:"level"`
	Confidence float64 `json:"confidence"`
}

type Profile struct {
	Name         string                   `json:"name"`
	DailyMinutes int                      `json:"dailyMinutes"`
	Goal         string                   `json:"goal"`
	Skills       map[string]SkillEstimate `json:"skills"`
	Interests    map[string]float64       `json:"interests"`
}

var SkillLabels = map[string]string{"python": "Python", "web": "Web 基础", "git": "Git", "algorithm": "算法基础"}
var InterestLabels = map[string]string{"backend": "后端开发", "frontend": "前端开发", "ai": "人工智能", "security": "网络安全"}

func unit(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 1 }

func (p Profile) Validate() error {
	if strings.TrimSpace(p.Name) == "" || utf8.RuneCountInString(p.Name) > 20 || p.DailyMinutes < 15 || p.DailyMinutes > 180 {
		return errors.New("昵称需为 1–20 个字符，每日学习时间需为 15–180 分钟")
	}
	if p.Goal != "project" && p.Goal != "opensource" && p.Goal != "competition" {
		return errors.New("请选择有效的学习目标")
	}
	if len(p.Skills) != len(SkillLabels) || len(p.Interests) != len(InterestLabels) {
		return errors.New("画像需包含全部已知技能和兴趣")
	}
	for id := range SkillLabels {
		v, ok := p.Skills[id]
		if !ok || !unit(v.Level) || !unit(v.Confidence) {
			return errors.New("能力及置信度需在 0–1 之间")
		}
	}
	for id := range InterestLabels {
		v, ok := p.Interests[id]
		if !ok || !unit(v) {
			return errors.New("兴趣需在 0–1 之间")
		}
	}
	return nil
}

type Task struct {
	ID             string             `json:"id"`
	Type           string             `json:"type"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	Category       string             `json:"category"`
	Tags           []string           `json:"tags"`
	RequiredSkills map[string]float64 `json:"requiredSkills"`
	Difficulty     float64            `json:"difficulty"`
	Minutes        int                `json:"minutes"`
	Quality        float64            `json:"quality"`
	Outcomes       []string           `json:"outcomes"`
	Steps          []string           `json:"steps"`
	Source         string             `json:"source"`
}

type Direction struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	English     string   `json:"english"`
	Description string   `json:"description"`
	Work        string   `json:"work"`
	Stack       []string `json:"stack"`
	Steps       []string `json:"steps"`
}

type Gap struct {
	Skill string  `json:"skill"`
	Gap   float64 `json:"gap"`
}
type Scores struct {
	Skill      float64 `json:"skill"`
	Interest   float64 `json:"interest"`
	Difficulty float64 `json:"difficulty"`
	Goal       float64 `json:"goal"`
	Quality    float64 `json:"quality"`
}
type Recommendation struct {
	Task          Task   `json:"task"`
	Score         int    `json:"score"`
	Scores        Scores `json:"scores"`
	Gaps          []Gap  `json:"gaps"`
	Reason        string `json:"reason"`
	Readiness     string `json:"readiness"`
	LowConfidence bool   `json:"lowConfidence"`
	RuleVersion   string `json:"ruleVersion"`
}

type Feedback struct {
	ID     string `json:"id"`
	TaskID string `json:"taskId"`
	Type   string `json:"type"`
}
