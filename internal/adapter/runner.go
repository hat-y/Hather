package adapter

import (
	"context"

	"github.com/hat-y/Hather/internal/model"
)

var order = []string{"macos", "ghostty", "herdr", "neovim", "vscode"}

type Runner struct{ Adapters []Adapter }

func (r Runner) Apply(ctx context.Context, apply ApplyContext, selected []string) model.OperationResult {
	adapters := make(map[string]Adapter, len(r.Adapters))
	for _, a := range r.Adapters {
		adapters[a.ID()] = a
	}
	result := model.OperationResult{Status: model.OperationComplete}
	for _, id := range requested(selected) {
		a, ok := adapters[id]
		if !ok {
			result.AdapterResults = append(result.AdapterResults, model.AdapterResult{AdapterID: id, Status: model.AdapterUnavailable, Diagnostics: []string{"adapter is not available in this build"}})
			result.Status = model.OperationPartialFailure
			continue
		}
		adapterResult := a.Apply(ctx, apply)
		adapterResult.AdapterID = id
		result.AdapterResults = append(result.AdapterResults, adapterResult)
		if adapterResult.Status != model.AdapterApplied {
			result.Status = model.OperationPartialFailure
		}
	}
	return result
}

func requested(selected []string) []string {
	seen := make(map[string]bool, len(selected))
	for _, id := range selected {
		seen[id] = true
	}
	out := make([]string, 0, len(selected))
	for _, id := range order {
		if seen[id] {
			out, seen[id] = append(out, id), false
		}
	}
	for _, id := range selected {
		if seen[id] {
			out, seen[id] = append(out, id), false
		}
	}
	return out
}
