package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/hat-y/Hather/internal/adapter"
	"github.com/hat-y/Hather/internal/config"
	"github.com/hat-y/Hather/internal/model"
)

type Source interface {
	Fetch(context.Context, model.Wallpaper) (model.Artifact, error)
}

type PaletteEngine interface {
	Extract(context.Context, model.Artifact) (model.Palette, error)
}

type Runner interface {
	Apply(context.Context, adapter.ApplyContext, []string) model.OperationResult
}

type ApplyRequest struct {
	Wallpaper model.Wallpaper
	Adapters  []string
}

type Service struct {
	Source Source
	Engine PaletteEngine
	Runner Runner
	Store  *config.Store
}

func (s Service) Search(ctx context.Context, query string) ([]model.Wallpaper, error) {
	searcher, ok := s.Source.(interface {
		Search(context.Context, string) ([]model.Wallpaper, error)
	})
	if !ok {
		return nil, fmt.Errorf("search is not configured")
	}
	return searcher.Search(ctx, query)
}

func (s Service) Preview(ctx context.Context, wallpaper model.Wallpaper) model.OperationResult {
	if s.Source == nil || s.Engine == nil {
		return prerequisiteFailure("application service is not configured")
	}
	artifact, err := s.Source.Fetch(ctx, wallpaper)
	if err != nil {
		return operationFailure(ctx, err)
	}
	palette, err := s.Engine.Extract(ctx, artifact)
	if err != nil {
		return operationFailure(ctx, err)
	}
	return model.OperationResult{Status: model.OperationComplete, Wallpaper: wallpaper, Artifact: artifact, Palette: palette}
}

func (s Service) Apply(ctx context.Context, request ApplyRequest) model.OperationResult {
	if s.Source == nil || s.Engine == nil || s.Runner == nil {
		return prerequisiteFailure("application service is not configured")
	}
	artifact, err := s.Source.Fetch(ctx, request.Wallpaper)
	if err != nil {
		return operationFailure(ctx, err)
	}
	if artifact.Path == "" {
		return prerequisiteFailure("wallpaper artifact has no local path")
	}
	palette, err := s.Engine.Extract(ctx, artifact)
	if err != nil {
		return operationFailure(ctx, err)
	}
	result := s.Runner.Apply(ctx, adapter.ApplyContext{Artifact: artifact, Palette: palette}, request.Adapters)
	result.Wallpaper, result.Artifact, result.Palette, result.RequestedAdapters = request.Wallpaper, artifact, palette, append([]string(nil), request.Adapters...)
	if s.Store != nil {
		if err := s.Store.SaveLastApplied(result); err != nil {
			result.Diagnostics = append(result.Diagnostics, "could not persist last-applied state: "+err.Error())
		}
	}
	return result
}

func operationFailure(ctx context.Context, err error) model.OperationResult {
	if ctx.Err() != nil || errors.Is(err, context.Canceled) {
		return model.OperationResult{Status: model.OperationCancelled, Diagnostics: []string{err.Error()}}
	}
	return prerequisiteFailure(err.Error())
}

func prerequisiteFailure(diagnostic string) model.OperationResult {
	return model.OperationResult{Status: model.OperationPrerequisiteFailure, Diagnostics: []string{diagnostic}}
}
