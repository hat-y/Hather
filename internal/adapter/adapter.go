package adapter

import (
	"context"

	"github.com/hat-y/Hather/internal/model"
)

// ApplyContext contains only the shared inputs adapters may need.
type ApplyContext struct {
	Artifact model.Artifact
	Palette  model.Palette
}

type Adapter interface {
	ID() string
	Apply(context.Context, ApplyContext) model.AdapterResult
}

type FuncAdapter struct {
	Name string
	Fn   func(context.Context, ApplyContext) model.AdapterResult
}

func (a FuncAdapter) ID() string { return a.Name }
func (a FuncAdapter) Apply(ctx context.Context, apply ApplyContext) model.AdapterResult {
	return a.Fn(ctx, apply)
}
