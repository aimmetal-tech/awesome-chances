package demo

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"awesome-chances/backend/internal/model"
)

//go:embed fixtures.json
var fixtures []byte

func Load() (model.Catalog, error) {
	var catalog model.Catalog
	if err := json.Unmarshal(fixtures, &catalog); err != nil {
		return catalog, err
	}
	if err := catalog.Profile.Validate(); err != nil {
		return catalog, err
	}
	ids := map[string]bool{}
	for _, task := range catalog.Tasks {
		if task.ID == "" || ids[task.ID] || task.Source != "demo" {
			return catalog, fmt.Errorf("invalid demo task %q", task.ID)
		}
		ids[task.ID] = true
		for skill := range task.RequiredSkills {
			if _, ok := model.SkillLabels[skill]; !ok {
				return catalog, fmt.Errorf("unknown skill %q", skill)
			}
		}
	}
	return catalog, nil
}
