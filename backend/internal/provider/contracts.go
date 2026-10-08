// Package provider defines replaceable external capabilities; it does not implement them.
package provider

import (
	"awesome-chances/backend/internal/model"
	"context"
)

type DecisionRequest struct {
	Kind    string
	Context map[string]any
}
type DecisionResult struct {
	Choice     string
	Scores     map[string]float64
	Confidence float64
}
type DecisionProvider interface {
	Decide(context.Context, DecisionRequest) (DecisionResult, error)
}
type GenerationProvider interface {
	Generate(context.Context, string) (string, error)
}
type EmbeddingProvider interface {
	Embed(context.Context, []string) ([][]float64, error)
}
type RepositoryProvider interface {
	SearchTasks(context.Context, string) ([]model.Task, error)
}
type CompetitionProvider interface {
	ListTasks(context.Context) ([]model.Task, error)
}
