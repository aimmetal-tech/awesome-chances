package model

import (
	"math"
	"testing"
)

func validFixture() Profile {
	p := Profile{Name: "测试者", DailyMinutes: 45, Goal: "project", Skills: map[string]SkillEstimate{}, Interests: map[string]float64{}}
	for id := range SkillLabels {
		p.Skills[id] = SkillEstimate{Level: .4, Confidence: .25}
	}
	for id := range InterestLabels {
		p.Interests[id] = .7
	}
	return p
}

func TestProfileValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		change func(*Profile)
	}{
		{"unknown skill", func(p *Profile) { p.Skills["unknown"] = SkillEstimate{} }},
		{"missing skill", func(p *Profile) { delete(p.Skills, "git") }},
		{"NaN interest", func(p *Profile) { p.Interests["ai"] = math.NaN() }},
		{"invalid confidence", func(p *Profile) { p.Skills["git"] = SkillEstimate{Level: .2, Confidence: 2} }},
		{"empty name", func(p *Profile) { p.Name = "  " }},
		{"invalid time", func(p *Profile) { p.DailyMinutes = 0 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := validFixture()
			test.change(&p)
			if p.Validate() == nil {
				t.Error("invalid profile accepted")
			}
		})
	}
	if err := validFixture().Validate(); err != nil {
		t.Fatal(err)
	}
}
